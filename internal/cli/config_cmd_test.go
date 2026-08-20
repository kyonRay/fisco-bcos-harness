package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConfigSetThenShowRoundTripsWithRedaction(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	t.Setenv("FBH_CONFIG", path)
	var stdout, stderr bytes.Buffer

	if code := Run([]string{"config", "set", "sheet_file_id", "SHEET123"}, &stdout, &stderr); code != 0 {
		t.Fatalf("config set sheet_file_id exit = %d, stderr: %s", code, stderr.String())
	}
	if code := Run([]string{"config", "set", "wecom_webhook", "https://qyapi.weixin.qq.com/hook?key=SECRETKEY"}, &stdout, &stderr); code != 0 {
		t.Fatalf("config set wecom_webhook exit = %d, stderr: %s", code, stderr.String())
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
	if !strings.Contains(out, "SHEET123") {
		t.Fatalf("show must print sheet_file_id, got: %s", out)
	}
	if strings.Contains(out, "SECRETKEY") {
		t.Fatalf("show must redact the webhook secret, got: %s", out)
	}
	if !strings.Contains(out, "wecom_webhook") {
		t.Fatalf("show must still name the wecom_webhook field, got: %s", out)
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
