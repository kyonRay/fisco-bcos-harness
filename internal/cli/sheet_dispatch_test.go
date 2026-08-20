package cli

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// recordsMCPServer serves list_records with the given existing records
// and captures add_records / update_records request arguments.
type recordsMCPServer struct {
	existing string // text content returned by list_records
	added    []map[string]any
	updated  []map[string]any
}

func (s *recordsMCPServer) start(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			ID     any    `json:"id"`
			Method string `json:"method"`
			Params struct {
				Name      string         `json:"name"`
				Arguments map[string]any `json:"arguments"`
			} `json:"params"`
		}
		json.NewDecoder(r.Body).Decode(&req)
		w.Header().Set("Content-Type", "application/json")
		reply := func(text string) {
			json.NewEncoder(w).Encode(map[string]any{
				"jsonrpc": "2.0", "id": req.ID,
				"result": map[string]any{"content": []map[string]any{{"type": "text", "text": text}}},
			})
		}
		switch req.Method {
		case "initialize":
			w.Header().Set("Mcp-Session-Id", "sess-1")
			json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": req.ID, "result": map[string]any{}})
		case "notifications/initialized":
			w.WriteHeader(http.StatusAccepted)
		case "tools/call":
			switch req.Params.Name {
			case "smartsheet.list_records":
				reply(s.existing)
			case "smartsheet.add_records":
				s.added = append(s.added, req.Params.Arguments)
				reply(`{"records":[{"record_id":"new1"}]}`)
			case "smartsheet.update_records":
				s.updated = append(s.updated, req.Params.Arguments)
				reply(`{"records":[{"record_id":"r1"}]}`)
			default:
				t.Errorf("unexpected tool %q", req.Params.Name)
			}
		}
	}))
}

func setupSheetEnv(t *testing.T, baseURL string) {
	t.Helper()
	cfgPath := filepath.Join(t.TempDir(), "config.json")
	t.Setenv("FBH_CONFIG", cfgPath)
	t.Setenv("FBH_MCPORTER_CONFIG", writeFakeMcporter(t, "Bearer FAKETOKEN"))
	cfg := `{"sheet_file_id":"SHEET123","mcp_base_url":"` + baseURL + `"}`
	if err := os.WriteFile(cfgPath, []byte(cfg), 0o600); err != nil {
		t.Fatal(err)
	}
}

const noRecords = `{"total":0,"records":[]}`

const loginRecordOwned = `{"total":1,"records":[{"record_id":"r1","field_values":[
  {"field":"需求名","text_value":{"items":[{"type":"text","text":"登录需求"}]}},
  {"field":"认领人","text_value":{"items":[{"type":"text","text":"李四"}]}}]}]}`

const loginRecordUnowned = `{"total":1,"records":[{"record_id":"r1","field_values":[
  {"field":"需求名","text_value":{"items":[{"type":"text","text":"登录需求"}]}}]}]}`

func TestUpsertRowAddsWhenKeyNotFound(t *testing.T) {
	srv := &recordsMCPServer{existing: noRecords}
	ts := srv.start(t)
	defer ts.Close()
	setupSheetEnv(t, ts.URL)
	var stdout, stderr bytes.Buffer

	code := Run([]string{"sheet", "upsert-row", "--table", "s1", "--key", "登录需求",
		"--set", "状态=待认领"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit = %d, stderr: %s", code, stderr.String())
	}
	if len(srv.added) != 1 || len(srv.updated) != 0 {
		t.Fatalf("added=%d updated=%d, want add path only", len(srv.added), len(srv.updated))
	}
	body, _ := json.Marshal(srv.added[0])
	for _, want := range []string{`"file_id":"SHEET123"`, `"sheet_id":"s1"`, `"登录需求"`, `"待认领"`, `"text_value"`} {
		if !strings.Contains(string(body), want) {
			t.Fatalf("add_records missing %s in %s", want, body)
		}
	}
}

func TestUpsertRowUpdatesWhenKeyFound(t *testing.T) {
	srv := &recordsMCPServer{existing: loginRecordUnowned}
	ts := srv.start(t)
	defer ts.Close()
	setupSheetEnv(t, ts.URL)
	var stdout, stderr bytes.Buffer

	code := Run([]string{"sheet", "set-status", "--table", "s1", "--key", "登录需求",
		"--status", "开发中"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit = %d, stderr: %s", code, stderr.String())
	}
	if len(srv.updated) != 1 || len(srv.added) != 0 {
		t.Fatalf("added=%d updated=%d, want update path only", len(srv.added), len(srv.updated))
	}
	body, _ := json.Marshal(srv.updated[0])
	for _, want := range []string{`"record_id":"r1"`, `"开发中"`} {
		if !strings.Contains(string(body), want) {
			t.Fatalf("update_records missing %s in %s", want, body)
		}
	}
}

func TestClaimRejectsWhenAlreadyOwnedNamingClaimant(t *testing.T) {
	srv := &recordsMCPServer{existing: loginRecordOwned}
	ts := srv.start(t)
	defer ts.Close()
	setupSheetEnv(t, ts.URL)
	var stdout, stderr bytes.Buffer

	code := Run([]string{"sheet", "claim", "--table", "s1", "--key", "登录需求", "--owner", "张三"}, &stdout, &stderr)

	if code != 1 {
		t.Fatalf("exit = %d, want 1; stdout: %s", code, stdout.String())
	}
	if !strings.Contains(stderr.String(), "李四") {
		t.Fatalf("stderr must name current claimant 李四, got: %s", stderr.String())
	}
	if len(srv.updated)+len(srv.added) != 0 {
		t.Fatalf("claim conflict must not write, added=%d updated=%d", len(srv.added), len(srv.updated))
	}
}

func TestClaimSetsOwnerAndStatusWhenUnowned(t *testing.T) {
	srv := &recordsMCPServer{existing: loginRecordUnowned}
	ts := srv.start(t)
	defer ts.Close()
	setupSheetEnv(t, ts.URL)
	var stdout, stderr bytes.Buffer

	code := Run([]string{"sheet", "claim", "--table", "s1", "--key", "登录需求", "--owner", "张三"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit = %d, stderr: %s", code, stderr.String())
	}
	if len(srv.updated) != 1 {
		t.Fatalf("updated=%d, want 1", len(srv.updated))
	}
	body, _ := json.Marshal(srv.updated[0])
	for _, want := range []string{`"张三"`, `"开发中"`, `"record_id":"r1"`} {
		if !strings.Contains(string(body), want) {
			t.Fatalf("claim update missing %s in %s", want, body)
		}
	}
}
