package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kyonRay/fisco-bcos-harness/internal/action"
)

// fullChainEnv wires stub gh + fake MCP together (requirement half in
// use, no PR ledger).
func fullChainEnv(t *testing.T) *recordsMCPServer {
	t.Helper()
	mcpSrv := &recordsMCPServer{existing: loginRecordUnowned}
	mcpTS := mcpSrv.start(t)
	t.Cleanup(mcpTS.Close)

	cfgPath := filepath.Join(t.TempDir(), "config.json")
	t.Setenv("FBH_CONFIG", cfgPath)
	t.Setenv("FBH_MCPORTER_CONFIG", writeFakeMcporter(t, "Bearer FAKETOKEN"))
	cfg := `{"sheet_file_id":"SHEET123","mcp_base_url":"` + mcpTS.URL + `"}`
	if err := os.WriteFile(cfgPath, []byte(cfg), 0o600); err != nil {
		t.Fatal(err)
	}
	writeGhStub(t, `case "$1 $2" in
"pr create") echo "https://github.com/team/repo/pull/42" ;;
*) echo "unexpected args: $*" >&2; exit 1 ;;
esac`)
	return mcpSrv
}

var prOpenArgs = []string{"pr", "open",
	"--title", "feat: login", "--body", "does login",
	"--reviewer", "lisi", "--table", "s1", "--key", "登录需求"}

func TestPrOpenDryRunEmitsTwoActionsNoSideEffects(t *testing.T) {
	mcpSrv := fullChainEnv(t)
	var stdout, stderr bytes.Buffer

	code := Run(append(prOpenArgs, "--dry-run"), &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit = %d, stderr: %s", code, stderr.String())
	}
	lines := strings.Split(strings.TrimSpace(stdout.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("dry-run must print exactly 2 action lines, got %d: %s", len(lines), stdout.String())
	}
	var ops []string
	for _, line := range lines {
		var a action.Action
		if err := json.Unmarshal([]byte(line), &a); err != nil {
			t.Fatalf("bad action line %q: %v", line, err)
		}
		ops = append(ops, a.Service+"."+a.Op)
	}
	want := "gh.create_pr sheet.upsert_row"
	if strings.Join(ops, " ") != want {
		t.Fatalf("ops = %v, want %s", ops, want)
	}
	if len(mcpSrv.added)+len(mcpSrv.updated) != 0 {
		t.Fatalf("dry-run must have zero side effects")
	}
}

func TestPrOpenFullChainCreatesAndWritesBack(t *testing.T) {
	mcpSrv := fullChainEnv(t)
	var stdout, stderr bytes.Buffer

	code := Run(prOpenArgs, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit = %d, stderr: %s", code, stderr.String())
	}
	if len(mcpSrv.updated) != 1 {
		t.Fatalf("sheet updated %d times, want 1", len(mcpSrv.updated))
	}
	sheetBody, _ := json.Marshal(mcpSrv.updated[0])
	for _, want := range []string{"https://github.com/team/repo/pull/42", "待review"} {
		if !strings.Contains(string(sheetBody), want) {
			t.Fatalf("sheet update missing %s in %s", want, sheetBody)
		}
	}
	if !strings.Contains(stdout.String(), "pull/42") {
		t.Fatalf("stdout must echo the created PR url, got: %s", stdout.String())
	}
}
