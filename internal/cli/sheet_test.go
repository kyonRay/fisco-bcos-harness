package cli

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestSheetPingUnconfiguredExitsOneAndPointsToSetup(t *testing.T) {
	t.Setenv("FBH_CONFIG", filepath.Join(t.TempDir(), "config.json")) // no such file
	var stdout, stderr bytes.Buffer

	code := Run([]string{"sheet", "ping"}, &stdout, &stderr)

	if code != 1 {
		t.Fatalf("exit code = %d, want 1; stdout: %s", code, stdout.String())
	}
	if !strings.Contains(stderr.String(), "fbh-setup") {
		t.Fatalf("stderr must point the user at /fbh-setup, got: %s", stderr.String())
	}
}
