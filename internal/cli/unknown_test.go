package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestUnknownCommandExitsTwoAndListsAvailable(t *testing.T) {
	registerFakeCmd(t, "fake-listed")
	var stdout, stderr bytes.Buffer

	code := Run([]string{"no-such-command"}, &stdout, &stderr)

	if code != 2 {
		t.Fatalf("exit code = %d, want 2", code)
	}
	msg := stderr.String()
	if !strings.Contains(msg, "no-such-command") {
		t.Fatalf("stderr does not name the bad command: %s", msg)
	}
	if !strings.Contains(msg, "fake-listed") {
		t.Fatalf("stderr does not list available commands: %s", msg)
	}
}
