package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ledgerEnv: PR-ledger mode — sheet configured with pr_sheet_id but the
// requirement half stays unused.
func ledgerEnv(t *testing.T, ghStubBody string) (*fakeWecom, *recordsMCPServer) {
	t.Helper()
	wecom := &fakeWecom{}
	wecomSrv := wecom.start(t)
	t.Cleanup(wecomSrv.Close)
	mcpSrv := &recordsMCPServer{existing: noRecords}
	mcpTS := mcpSrv.start(t)
	t.Cleanup(mcpTS.Close)

	cfgPath := filepath.Join(t.TempDir(), "config.json")
	t.Setenv("FBH_CONFIG", cfgPath)
	t.Setenv("FBH_MCPORTER_CONFIG", writeFakeMcporter(t, "Bearer FAKETOKEN"))
	cfgJSON, _ := json.Marshal(map[string]string{
		"sheet_file_id": "SHEET123",
		"pr_sheet_id":   "prlog",
		"wecom_webhook": wecomSrv.URL,
		"mcp_base_url":  mcpTS.URL,
	})
	if err := os.WriteFile(cfgPath, cfgJSON, 0o600); err != nil {
		t.Fatal(err)
	}
	writeGhStub(t, ghStubBody)
	return wecom, mcpSrv
}

func TestPrOpenRegistersInPRLedger(t *testing.T) {
	wecom, mcpSrv := ledgerEnv(t, `case "$1 $2" in
"pr create") echo "https://github.com/t/r/pull/9" ;;
"api user") echo "zhangsan" ;;
*) echo "unexpected: $*" >&2; exit 1 ;;
esac`)
	var stdout, stderr bytes.Buffer

	code := Run([]string{"pr", "open", "--title", "feat: y", "--reviewer", "lisi,wangwu"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit = %d, stderr: %s", code, stderr.String())
	}
	body, _ := json.Marshal(mcpSrv.added)
	for _, want := range []string{`"sheet_id":"prlog"`, "pull/9", "待review", "lisi,wangwu", "zhangsan", "feat: y"} {
		if !strings.Contains(string(body), want) {
			t.Fatalf("ledger row missing %q in %s", want, body)
		}
	}
	if len(wecom.bodies) != 1 {
		t.Fatalf("wecom sent %d, want 1", len(wecom.bodies))
	}
	var msg struct {
		Text struct {
			MentionedList []string `json:"mentioned_list"`
		} `json:"text"`
	}
	json.Unmarshal([]byte(wecom.bodies[0]), &msg)
	if len(msg.Text.MentionedList) != 2 {
		t.Fatalf("both reviewers must be mentioned, got %v", msg.Text.MentionedList)
	}
}

func syncViewStub(fixtureEnv string) string {
	return fmt.Sprintf(`case "$1 $2" in
"pr view") echo "$%s" ;;
*) echo "unexpected: $*" >&2; exit 1 ;;
esac`, fixtureEnv)
}

func TestPrSyncFixedPRBecomesPendingRereviewAndNotifiesOnlyUnapproved(t *testing.T) {
	// lisi request-changed 20h ago, wangwu approved; author pushed 5h ago.
	fixture := fmt.Sprintf(`{"url":"https://github.com/t/r/pull/9","title":"feat: y","state":"OPEN",
"author":{"login":"zhangsan"},"createdAt":%q,"updatedAt":%q,"reviewDecision":"CHANGES_REQUESTED",
"reviewRequests":[],"latestReviews":[
 {"author":{"login":"lisi"},"state":"CHANGES_REQUESTED","submittedAt":%q},
 {"author":{"login":"wangwu"},"state":"APPROVED","submittedAt":%q}]}`,
		ago(50), ago(5), ago(20), ago(10))
	t.Setenv("FBH_TEST_VIEW", fixture)
	wecom, mcpSrv := ledgerEnv(t, syncViewStub("FBH_TEST_VIEW"))
	var stdout, stderr bytes.Buffer

	code := Run([]string{"pr", "sync", "--pr", "https://github.com/t/r/pull/9", "--notify"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit = %d, stderr: %s", code, stderr.String())
	}
	body, _ := json.Marshal(mcpSrv.added)
	for _, want := range []string{"待复审", "wangwu"} { // 已approve 列记 wangwu
		if !strings.Contains(string(body), want) {
			t.Fatalf("synced row missing %q in %s", want, body)
		}
	}
	if len(wecom.bodies) != 1 {
		t.Fatalf("wecom sent %d, want 1", len(wecom.bodies))
	}
	if !strings.Contains(wecom.bodies[0], "lisi") || strings.Contains(wecom.bodies[0], `"wangwu"`) {
		t.Fatalf("must mention only the un-approved lisi, got: %s", wecom.bodies[0])
	}
}

func TestPrSyncApprovedPRNotifiesNobody(t *testing.T) {
	fixture := fmt.Sprintf(`{"url":"https://github.com/t/r/pull/9","title":"feat: y","state":"OPEN",
"author":{"login":"zhangsan"},"createdAt":%q,"updatedAt":%q,"reviewDecision":"APPROVED",
"reviewRequests":[],"latestReviews":[{"author":{"login":"lisi"},"state":"APPROVED","submittedAt":%q}]}`,
		ago(50), ago(5), ago(6))
	t.Setenv("FBH_TEST_VIEW", fixture)
	wecom, mcpSrv := ledgerEnv(t, syncViewStub("FBH_TEST_VIEW"))
	var stdout, stderr bytes.Buffer

	code := Run([]string{"pr", "sync", "--pr", "https://github.com/t/r/pull/9", "--notify"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit = %d, stderr: %s", code, stderr.String())
	}
	body, _ := json.Marshal(mcpSrv.added)
	if !strings.Contains(string(body), "已approve") {
		t.Fatalf("synced row must be 已approve, got %s", body)
	}
	if len(wecom.bodies) != 0 {
		t.Fatalf("approved PR must notify nobody, sent %d", len(wecom.bodies))
	}
}

const ledgerRows = `{"total":2,"records":[
{"record_id":"l1","field_values":[
  {"field":"PR链接","text_value":{"items":[{"type":"text","text":"https://github.com/t/r/pull/9"}]}},
  {"field":"状态","text_value":{"items":[{"type":"text","text":"待review"}]}},
  {"field":"reviewers","text_value":{"items":[{"type":"text","text":"lisi"}]}}]},
{"record_id":"l2","field_values":[
  {"field":"PR链接","text_value":{"items":[{"type":"text","text":"https://github.com/t/r/pull/8"}]}},
  {"field":"状态","text_value":{"items":[{"type":"text","text":"已合入"}]}}]}]}`

func TestPrBoardListsOnlyOpenWork(t *testing.T) {
	_, mcpSrv := ledgerEnv(t, `echo "unexpected: $*" >&2; exit 1`)
	mcpSrv.existing = ledgerRows
	var stdout, stderr bytes.Buffer

	code := Run([]string{"pr", "board"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit = %d, stderr: %s", code, stderr.String())
	}
	out := stdout.String()
	if !strings.Contains(out, "pull/9") || !strings.Contains(out, "待review") {
		t.Fatalf("board must list the open PR, got: %s", out)
	}
	if strings.Contains(out, "pull/8") {
		t.Fatalf("board must hide 已合入 rows, got: %s", out)
	}
}
