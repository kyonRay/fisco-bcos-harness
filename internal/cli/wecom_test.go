package cli

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// fakeWecom captures webhook POST bodies.
type fakeWecom struct {
	bodies []string
}

func (f *fakeWecom) start(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		f.bodies = append(f.bodies, string(b))
		w.Write([]byte(`{"errcode":0}`))
	}))
}

func configureWecom(t *testing.T, webhookURL string) {
	t.Helper()
	cfgPath := filepath.Join(t.TempDir(), "config.json")
	t.Setenv("FBH_CONFIG", cfgPath)
	cfg := `{"wecom_webhook":"` + webhookURL + `","sheet_file_id":"SHEET123"}`
	if err := os.WriteFile(cfgPath, []byte(cfg), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestWecomNudgeSendsDirectedMention(t *testing.T) {
	wecom := &fakeWecom{}
	srv := wecom.start(t)
	defer srv.Close()
	configureWecom(t, srv.URL)
	var stdout, stderr bytes.Buffer

	code := Run([]string{"wecom", "nudge", "--to", "zhangsan",
		"--pr", "https://github.com/team/repo/pull/42",
		"--text", "请 review"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit = %d, stderr: %s", code, stderr.String())
	}
	if len(wecom.bodies) != 1 {
		t.Fatalf("webhook called %d times, want exactly 1", len(wecom.bodies))
	}
	var msg struct {
		MsgType string `json:"msgtype"`
		Text    struct {
			Content       string   `json:"content"`
			MentionedList []string `json:"mentioned_list"`
		} `json:"text"`
	}
	if err := json.Unmarshal([]byte(wecom.bodies[0]), &msg); err != nil {
		t.Fatalf("bad webhook body: %v", err)
	}
	if msg.MsgType != "text" {
		t.Fatalf("msgtype = %q", msg.MsgType)
	}
	if len(msg.Text.MentionedList) != 1 || msg.Text.MentionedList[0] != "zhangsan" {
		t.Fatalf("mentioned_list = %v, want exactly [zhangsan]", msg.Text.MentionedList)
	}
	if !strings.Contains(msg.Text.Content, "https://github.com/team/repo/pull/42") {
		t.Fatalf("content must carry the PR link, got: %s", msg.Text.Content)
	}
}

func TestWecomNudgeUnconfiguredPointsToSetup(t *testing.T) {
	t.Setenv("FBH_CONFIG", filepath.Join(t.TempDir(), "config.json"))
	var stdout, stderr bytes.Buffer

	code := Run([]string{"wecom", "nudge", "--to", "x", "--pr", "y"}, &stdout, &stderr)

	if code != 1 {
		t.Fatalf("exit = %d, want 1", code)
	}
	if !strings.Contains(stderr.String(), "fbh-setup") {
		t.Fatalf("stderr must point at /fbh-setup, got: %s", stderr.String())
	}
}
