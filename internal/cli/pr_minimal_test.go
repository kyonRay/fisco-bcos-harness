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

// minimalEnv: wecom + gh only — no sheet_file_id, no MCP server at all.
// This is the "只治 review" adoption mode.
func minimalEnv(t *testing.T) *fakeWecom {
	t.Helper()
	wecom := &fakeWecom{}
	srv := wecom.start(t)
	t.Cleanup(srv.Close)
	cfgPath := filepath.Join(t.TempDir(), "config.json")
	t.Setenv("FBH_CONFIG", cfgPath)
	if err := os.WriteFile(cfgPath, []byte(`{"wecom_webhook":"`+srv.URL+`"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	writeGhStub(t, `case "$1 $2" in
"pr create") echo "https://github.com/team/repo/pull/7" ;;
*) echo "unexpected args: $*" >&2; exit 1 ;;
esac`)
	return wecom
}

var minimalPrArgs = []string{"pr", "open", "--title", "feat: x", "--reviewer", "lisi"}

func TestPrOpenWithoutSheetConfigRunsTwoStepChain(t *testing.T) {
	wecom := minimalEnv(t)
	var stdout, stderr bytes.Buffer

	code := Run(minimalPrArgs, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit = %d, stderr: %s", code, stderr.String())
	}
	if len(wecom.bodies) != 1 || !strings.Contains(wecom.bodies[0], "pull/7") {
		t.Fatalf("wecom must be nudged with the PR url, bodies: %v", wecom.bodies)
	}
	if !strings.Contains(stdout.String(), "pull/7") {
		t.Fatalf("stdout must echo the created PR url, got: %s", stdout.String())
	}
}

func TestPrOpenWithoutSheetDryRunEmitsTwoActions(t *testing.T) {
	wecom := minimalEnv(t)
	var stdout, stderr bytes.Buffer

	code := Run(append(minimalPrArgs, "--dry-run"), &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit = %d, stderr: %s", code, stderr.String())
	}
	var ops []string
	for _, line := range strings.Split(strings.TrimSpace(stdout.String()), "\n") {
		if !strings.HasPrefix(line, "{") {
			continue
		}
		var a action.Action
		if err := json.Unmarshal([]byte(line), &a); err != nil {
			t.Fatalf("bad action line %q: %v", line, err)
		}
		ops = append(ops, a.Service+"."+a.Op)
	}
	if strings.Join(ops, " ") != "gh.create_pr wecom.nudge" {
		t.Fatalf("ops = %v, want [gh.create_pr wecom.nudge] (no sheet action)", ops)
	}
	if len(wecom.bodies) != 0 {
		t.Fatalf("dry-run must not send")
	}
}

func TestPrOpenTableWithoutSheetConfigErrorsToSetup(t *testing.T) {
	minimalEnv(t)
	var stdout, stderr bytes.Buffer

	code := Run(append(minimalPrArgs, "--table", "s1", "--key", "需求"), &stdout, &stderr)

	if code != 1 {
		t.Fatalf("exit = %d, want 1 (--table given but sheet unconfigured)", code)
	}
	if !strings.Contains(stderr.String(), "fbh-setup") {
		t.Fatalf("stderr must point at /fbh-setup, got: %s", stderr.String())
	}
}
