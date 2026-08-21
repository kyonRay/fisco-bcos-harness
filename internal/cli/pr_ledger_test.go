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
func ledgerEnv(t *testing.T, ghStubBody string) *recordsMCPServer {
	t.Helper()
	mcpSrv := &recordsMCPServer{existing: noRecords}
	mcpTS := mcpSrv.start(t)
	t.Cleanup(mcpTS.Close)

	cfgPath := filepath.Join(t.TempDir(), "config.json")
	t.Setenv("FBH_CONFIG", cfgPath)
	t.Setenv("FBH_MCPORTER_CONFIG", writeFakeMcporter(t, "Bearer FAKETOKEN"))
	cfgJSON, _ := json.Marshal(map[string]string{
		"sheet_file_id": "SHEET123",
		"pr_sheet_id":   "prlog",
		"mcp_base_url":  mcpTS.URL,
	})
	if err := os.WriteFile(cfgPath, cfgJSON, 0o600); err != nil {
		t.Fatal(err)
	}
	writeGhStub(t, ghStubBody)
	return mcpSrv
}

// capturedField digs one column's text out of a captured add/update
// request (the smartsheet argument shape).
func capturedField(t *testing.T, args map[string]any, field string) (string, bool) {
	t.Helper()
	records, _ := args["records"].([]any)
	for _, r := range records {
		rm, _ := r.(map[string]any)
		fvs, _ := rm["field_values"].([]any)
		for _, fv := range fvs {
			fvm, _ := fv.(map[string]any)
			if fvm["field"] != field {
				continue
			}
			tv, _ := fvm["text_value"].(map[string]any)
			items, _ := tv["items"].([]any)
			var sb strings.Builder
			for _, it := range items {
				im, _ := it.(map[string]any)
				s, _ := im["text"].(string)
				sb.WriteString(s)
			}
			return sb.String(), true
		}
	}
	return "", false
}

func TestPrOpenRegistersInPRLedgerWithPendingReviewers(t *testing.T) {
	mcpSrv := ledgerEnv(t, `case "$1 $2" in
"pr create") echo "https://github.com/t/r/pull/9" ;;
"api user") echo "zhangsan" ;;
*) echo "unexpected: $*" >&2; exit 1 ;;
esac`)
	var stdout, stderr bytes.Buffer

	code := Run([]string{"pr", "open", "--title", "feat: y", "--reviewer", "lisi,wangwu"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit = %d, stderr: %s", code, stderr.String())
	}
	if len(mcpSrv.added) != 1 {
		t.Fatalf("ledger rows added = %d, want 1", len(mcpSrv.added))
	}
	body, _ := json.Marshal(mcpSrv.added)
	for _, want := range []string{`"sheet_id":"prlog"`, "pull/9", "待review", "zhangsan", "feat: y"} {
		if !strings.Contains(string(body), want) {
			t.Fatalf("ledger row missing %q in %s", want, body)
		}
	}
	// 待处理人 must start as the full reviewer list.
	if got, ok := capturedField(t, mcpSrv.added[0], "待处理人"); !ok || got != "lisi,wangwu" {
		t.Fatalf("待处理人 = %q (found=%v), want lisi,wangwu", got, ok)
	}
	// The paste-ready group broadcast names the PR and its reviewers.
	out := stdout.String()
	for _, want := range []string{"复制到企微群", "pull/9", "lisi,wangwu", "smartsheet/SHEET123"} {
		if !strings.Contains(out, want) {
			t.Fatalf("broadcast missing %q, got:\n%s", want, out)
		}
	}
}

func syncViewStub(fixtureEnv string) string {
	return fmt.Sprintf(`case "$1 $2" in
"pr view") echo "$%s" ;;
*) echo "unexpected: $*" >&2; exit 1 ;;
esac`, fixtureEnv)
}

func TestPrSyncFixedPRBecomesPendingRereviewWithOnlyUnapprovedPending(t *testing.T) {
	// lisi request-changed 20h ago, wangwu approved; author pushed 5h ago.
	fixture := fmt.Sprintf(`{"url":"https://github.com/t/r/pull/9","title":"feat: y","state":"OPEN",
"author":{"login":"zhangsan"},"createdAt":%q,"updatedAt":%q,"reviewDecision":"CHANGES_REQUESTED",
"reviewRequests":[],"latestReviews":[
 {"author":{"login":"lisi"},"state":"CHANGES_REQUESTED","submittedAt":%q},
 {"author":{"login":"wangwu"},"state":"APPROVED","submittedAt":%q}]}`,
		ago(50), ago(5), ago(20), ago(10))
	t.Setenv("FBH_TEST_VIEW", fixture)
	mcpSrv := ledgerEnv(t, syncViewStub("FBH_TEST_VIEW"))
	var stdout, stderr bytes.Buffer

	code := Run([]string{"pr", "sync", "--pr", "https://github.com/t/r/pull/9"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit = %d, stderr: %s", code, stderr.String())
	}
	if len(mcpSrv.added) != 1 {
		t.Fatalf("ledger rows added = %d, want 1", len(mcpSrv.added))
	}
	if got, _ := capturedField(t, mcpSrv.added[0], "状态"); got != "待复审" {
		t.Fatalf("状态 = %q, want 待复审", got)
	}
	if got, _ := capturedField(t, mcpSrv.added[0], "已approve"); got != "wangwu" {
		t.Fatalf("已approve = %q, want wangwu", got)
	}
	// Only the un-approved lisi stays pending; wangwu must NOT
	// reappear here.
	if got, _ := capturedField(t, mcpSrv.added[0], "待处理人"); got != "lisi" {
		t.Fatalf("待处理人 = %q, want lisi", got)
	}
	// Broadcast asks only lisi to continue; the approved wangwu is not
	// on the 还差 line.
	out := stdout.String()
	for _, want := range []string{"复制到企微群", "请继续 review", "还差 approve: lisi"} {
		if !strings.Contains(out, want) {
			t.Fatalf("broadcast missing %q, got:\n%s", want, out)
		}
	}
	if strings.Contains(out, "还差 approve: lisi,wangwu") || strings.Contains(out, "wangwu,lisi") {
		t.Fatalf("approved wangwu must not be asked again, got:\n%s", out)
	}
}

func TestPrSyncApprovedPRClearsPending(t *testing.T) {
	fixture := fmt.Sprintf(`{"url":"https://github.com/t/r/pull/9","title":"feat: y","state":"OPEN",
"author":{"login":"zhangsan"},"createdAt":%q,"updatedAt":%q,"reviewDecision":"APPROVED",
"reviewRequests":[],"latestReviews":[{"author":{"login":"lisi"},"state":"APPROVED","submittedAt":%q}]}`,
		ago(50), ago(5), ago(6))
	t.Setenv("FBH_TEST_VIEW", fixture)
	mcpSrv := ledgerEnv(t, syncViewStub("FBH_TEST_VIEW"))
	var stdout, stderr bytes.Buffer

	code := Run([]string{"pr", "sync", "--pr", "https://github.com/t/r/pull/9"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit = %d, stderr: %s", code, stderr.String())
	}
	if got, _ := capturedField(t, mcpSrv.added[0], "状态"); got != "已approve" {
		t.Fatalf("状态 = %q, want 已approve", got)
	}
	if got, ok := capturedField(t, mcpSrv.added[0], "待处理人"); !ok || got != "" {
		t.Fatalf("待处理人 = %q (found=%v), want cleared — nobody left to remind", got, ok)
	}
	if strings.Contains(stdout.String(), "复制到企微群") {
		t.Fatalf("approved PR must not draft a broadcast, got:\n%s", stdout.String())
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
	mcpSrv := ledgerEnv(t, `echo "unexpected: $*" >&2; exit 1`)
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
