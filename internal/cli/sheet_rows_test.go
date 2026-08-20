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

func configureSheet(t *testing.T) {
	t.Helper()
	cfgPath := filepath.Join(t.TempDir(), "config.json")
	t.Setenv("FBH_CONFIG", cfgPath)
	if err := os.WriteFile(cfgPath, []byte(`{"sheet_file_id":"SHEET123"}`), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestSetStatusRejectsUnknownStatusListingEnum(t *testing.T) {
	configureSheet(t)
	var stdout, stderr bytes.Buffer

	code := Run([]string{"sheet", "set-status", "--table", "s1", "--key", "登录需求", "--status", "瞎写的"}, &stdout, &stderr)

	if code != 1 {
		t.Fatalf("exit = %d, want 1", code)
	}
	msg := stderr.String()
	if !strings.Contains(msg, "瞎写的") || !strings.Contains(msg, "待认领") || !strings.Contains(msg, "完成") {
		t.Fatalf("stderr must name the bad status and list the enum, got: %s", msg)
	}
}

func TestSetStatusDryRunEmitsUpsertAction(t *testing.T) {
	configureSheet(t)
	var stdout, stderr bytes.Buffer

	code := Run([]string{"sheet", "set-status", "--table", "s1", "--key", "登录需求", "--status", "开发中", "--dry-run"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit = %d, stderr: %s", code, stderr.String())
	}
	var a action.Action
	if err := json.Unmarshal([]byte(strings.TrimSpace(stdout.String())), &a); err != nil {
		t.Fatalf("stdout not one action JSON line: %q: %v", stdout.String(), err)
	}
	if a.Service != "sheet" || a.Op != "upsert_row" {
		t.Fatalf("action = %+v, want sheet.upsert_row", a)
	}
	sets, _ := a.Payload["sets"].(map[string]any)
	if a.Payload["sheet_id"] != "s1" || a.Payload["key"] != "登录需求" || sets["状态"] != "开发中" {
		t.Fatalf("payload = %v", a.Payload)
	}
}

func TestUpsertRowDryRunEmitsAllSets(t *testing.T) {
	configureSheet(t)
	var stdout, stderr bytes.Buffer

	code := Run([]string{"sheet", "upsert-row", "--table", "s1", "--key", "登录需求",
		"--set", "milestone=M1", "--set", "认领人=张三", "--set", "状态=待认领", "--dry-run"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit = %d, stderr: %s", code, stderr.String())
	}
	var a action.Action
	if err := json.Unmarshal([]byte(strings.TrimSpace(stdout.String())), &a); err != nil {
		t.Fatalf("stdout not one action JSON line: %v", err)
	}
	sets, _ := a.Payload["sets"].(map[string]any)
	if sets["milestone"] != "M1" || sets["认领人"] != "张三" || sets["状态"] != "待认领" {
		t.Fatalf("sets = %v", sets)
	}
	if a.Payload["key_field"] != "需求名" {
		t.Fatalf("key_field = %v, want 需求名 (schema default)", a.Payload["key_field"])
	}
}

func TestUpsertRowRejectsInvalidStatusInSets(t *testing.T) {
	configureSheet(t)
	var stdout, stderr bytes.Buffer

	code := Run([]string{"sheet", "upsert-row", "--table", "s1", "--key", "x",
		"--set", "状态=乱来", "--dry-run"}, &stdout, &stderr)

	if code != 1 {
		t.Fatalf("exit = %d, want 1", code)
	}
	if !strings.Contains(stderr.String(), "乱来") {
		t.Fatalf("stderr must name the bad status, got: %s", stderr.String())
	}
}
