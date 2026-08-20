package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kyonRay/fisco-bcos-harness/internal/action"
)

// Two requirement rows: one in M1, one in M2. complete_milestone must
// touch only the M1 row.
const milestoneRecords = `{"total":2,"records":[
{"record_id":"r1","field_values":[
  {"field":"需求名","text_value":{"items":[{"type":"text","text":"登录需求"}]}},
  {"field":"milestone","text_value":{"items":[{"type":"text","text":"M1"}]}}]},
{"record_id":"r2","field_values":[
  {"field":"需求名","text_value":{"items":[{"type":"text","text":"支付需求"}]}},
  {"field":"milestone","text_value":{"items":[{"type":"text","text":"M2"}]}}]}]}`

func gateEnv(t *testing.T, gateCmd string) (*fakeWecom, *recordsMCPServer) {
	t.Helper()
	wecom := &fakeWecom{}
	wecomSrv := wecom.start(t)
	t.Cleanup(wecomSrv.Close)
	mcpSrv := &recordsMCPServer{existing: milestoneRecords}
	mcpTS := mcpSrv.start(t)
	t.Cleanup(mcpTS.Close)

	cfgPath := filepath.Join(t.TempDir(), "config.json")
	t.Setenv("FBH_CONFIG", cfgPath)
	t.Setenv("FBH_MCPORTER_CONFIG", writeFakeMcporter(t, "Bearer FAKETOKEN"))
	cfgJSON, _ := json.Marshal(map[string]string{
		"sheet_file_id": "SHEET123",
		"wecom_webhook": wecomSrv.URL,
		"mcp_base_url":  mcpTS.URL,
		"gate_cmd":      gateCmd,
	})
	if err := os.WriteFile(cfgPath, cfgJSON, 0o600); err != nil {
		t.Fatal(err)
	}
	return wecom, mcpSrv
}

var gateArgs = []string{"gate", "run", "--milestone", "M1",
	"--milestone-table", "ms", "--reqs-table", "s1", "--owner", "wangwu"}

func TestGateSuccessDryRunEmitsWriteBackSequence(t *testing.T) {
	wecom, mcpSrv := gateEnv(t, "exit 0")
	var stdout, stderr bytes.Buffer

	code := Run(append(gateArgs, "--dry-run"), &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit = %d, stderr: %s", code, stderr.String())
	}
	var ops []string
	for _, line := range strings.Split(strings.TrimSpace(stdout.String()), "\n") {
		if !strings.HasPrefix(line, "{") {
			continue // skip human status lines
		}
		var a action.Action
		if err := json.Unmarshal([]byte(line), &a); err != nil {
			t.Fatalf("bad action line %q: %v", line, err)
		}
		ops = append(ops, a.Service+"."+a.Op)
	}
	want := "sheet.upsert_row sheet.complete_milestone"
	if strings.Join(ops, " ") != want {
		t.Fatalf("ops = %v, want %s", ops, want)
	}
	if len(wecom.bodies) != 0 || len(mcpSrv.added)+len(mcpSrv.updated) != 0 {
		t.Fatalf("dry-run must have zero side effects")
	}
}

func TestGateSuccessWritesBackAndCompletesOnlyThisMilestone(t *testing.T) {
	_, mcpSrv := gateEnv(t, "exit 0")
	var stdout, stderr bytes.Buffer

	code := Run(gateArgs, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit = %d, stderr: %s", code, stderr.String())
	}
	addBody, _ := json.Marshal(mcpSrv.added)
	if !strings.Contains(string(addBody), "通过") {
		t.Fatalf("milestone row must record 门禁状态=通过, added: %s", addBody)
	}
	updBody, _ := json.Marshal(mcpSrv.updated)
	if !strings.Contains(string(updBody), `"record_id":"r1"`) || !strings.Contains(string(updBody), "完成") {
		t.Fatalf("M1 requirement must be completed, updated: %s", updBody)
	}
	if strings.Contains(string(updBody), `"record_id":"r2"`) {
		t.Fatalf("M2 requirement must NOT be touched, updated: %s", updBody)
	}
}

func TestGateFailureCreatesDefectRowAndNudgesFallbackOwner(t *testing.T) {
	wecom, mcpSrv := gateEnv(t, "echo consensus stall detected; exit 1")
	var stdout, stderr bytes.Buffer

	code := Run(gateArgs, &stdout, &stderr)

	if code != 1 {
		t.Fatalf("gate failure must exit 1, got %d; stderr: %s", code, stderr.String())
	}
	addBody, _ := json.Marshal(mcpSrv.added)
	for _, want := range []string{"gate-defect", "待认领", "M1", "consensus stall detected"} {
		if !strings.Contains(string(addBody), want) {
			t.Fatalf("defect row missing %q, added: %s", want, addBody)
		}
	}
	if strings.Contains(string(addBody), "完成") {
		t.Fatalf("failure path must not complete requirements, added: %s", addBody)
	}
	// r1/r2 状态不被误改：唯一允许的写是缺陷行 add。
	if len(mcpSrv.updated) != 0 {
		t.Fatalf("failure path must not update existing rows, updated: %d", len(mcpSrv.updated))
	}
	if len(wecom.bodies) != 1 || !strings.Contains(wecom.bodies[0], "wangwu") {
		t.Fatalf("failure must nudge fallback owner wangwu once, bodies: %v", wecom.bodies)
	}
}

func TestGateFailureNudgesExplicitAssignee(t *testing.T) {
	wecom, _ := gateEnv(t, "exit 1")
	var stdout, stderr bytes.Buffer

	code := Run(append(gateArgs, "--assignee", "zhaoliu"), &stdout, &stderr)

	if code != 1 {
		t.Fatalf("exit = %d, want 1", code)
	}
	if len(wecom.bodies) != 1 || !strings.Contains(wecom.bodies[0], "zhaoliu") {
		t.Fatalf("must nudge attributed author zhaoliu, bodies: %v", wecom.bodies)
	}
}

func TestGateWithoutGateCmdPointsToConfig(t *testing.T) {
	gateEnv(t, "")
	var stdout, stderr bytes.Buffer

	code := Run(gateArgs, &stdout, &stderr)

	if code != 1 {
		t.Fatalf("exit = %d, want 1", code)
	}
	if !strings.Contains(stderr.String(), "gate_cmd") {
		t.Fatalf("stderr must mention gate_cmd config, got: %s", stderr.String())
	}
}

var _ = fmt.Sprintf // keep fmt import if assertions change
