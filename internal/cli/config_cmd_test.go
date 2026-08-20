package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConfigSetThenShowRoundTrips(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	t.Setenv("FBH_CONFIG", path)
	var stdout, stderr bytes.Buffer

	if code := Run([]string{"config", "set", "sheet_file_id", "SHEET123"}, &stdout, &stderr); code != 0 {
		t.Fatalf("config set sheet_file_id exit = %d, stderr: %s", code, stderr.String())
	}
	if code := Run([]string{"config", "set", "pr_sheet_id", "prlog"}, &stdout, &stderr); code != 0 {
		t.Fatalf("config set pr_sheet_id exit = %d, stderr: %s", code, stderr.String())
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("config file not written: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Fatalf("config file mode = %o, want 600", perm)
	}

	stdout.Reset()
	if code := Run([]string{"config", "show"}, &stdout, &stderr); code != 0 {
		t.Fatalf("config show exit = %d, stderr: %s", code, stderr.String())
	}
	out := stdout.String()
	for _, want := range []string{"SHEET123", "prlog"} {
		if !strings.Contains(out, want) {
			t.Fatalf("show must print %s, got: %s", want, out)
		}
	}
}

func TestConfigSetRejectsRemovedWecomKeys(t *testing.T) {
	t.Setenv("FBH_CONFIG", filepath.Join(t.TempDir(), "config.json"))
	for _, key := range []string{"wecom_webhook", "mention_map"} {
		var stdout, stderr bytes.Buffer
		if code := Run([]string{"config", "set", key, "v"}, &stdout, &stderr); code != 1 {
			t.Fatalf("set %s exit = %d, want 1 (key removed: sheet automation notifies now)", key, code)
		}
	}
}

func TestConfigSetRejectsUnknownKey(t *testing.T) {
	t.Setenv("FBH_CONFIG", filepath.Join(t.TempDir(), "config.json"))
	var stdout, stderr bytes.Buffer

	code := Run([]string{"config", "set", "no_such_key", "v"}, &stdout, &stderr)

	if code != 1 {
		t.Fatalf("exit = %d, want 1", code)
	}
	if !strings.Contains(stderr.String(), "no_such_key") {
		t.Fatalf("stderr must name the bad key, got: %s", stderr.String())
	}
}
