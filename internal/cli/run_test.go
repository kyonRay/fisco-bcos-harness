package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestVersionFlagPrintsVersionAndExitsZero(t *testing.T) {
	var stdout, stderr bytes.Buffer

	code := Run([]string{"--version"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", code, stderr.String())
	}
	got := strings.TrimSpace(stdout.String())
	want := "fbh " + Version
	if got != want {
		t.Fatalf("stdout = %q, want %q", got, want)
	}
}
