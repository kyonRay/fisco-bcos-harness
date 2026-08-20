package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeGhStub installs a fake gh binary whose behavior is driven by a
// case statement over its arguments; body is the script fragment.
func writeGhStub(t *testing.T, body string) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "gh")
	script := "#!/bin/bash\n" + body + "\n"
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("FBH_GH_BIN", path)
}

const reviewStateStub = `case "$*" in
*pull/41*) echo '{"reviewDecision":"APPROVED"}' ;;
*pull/42*) echo '{"reviewDecision":"CHANGES_REQUESTED"}' ;;
*pull/43*) echo '{"reviewDecision":""}' ;;
*) echo "unexpected args: $*" >&2; exit 1 ;;
esac`

func TestGhReviewStateThreeStates(t *testing.T) {
	writeGhStub(t, reviewStateStub)
	cases := []struct{ pr, want string }{
		{"https://github.com/team/repo/pull/41", "approved"},
		{"https://github.com/team/repo/pull/42", "changes_requested"},
		{"https://github.com/team/repo/pull/43", "none"},
	}
	for _, tc := range cases {
		var stdout, stderr bytes.Buffer
		code := Run([]string{"gh", "review-state", "--pr", tc.pr}, &stdout, &stderr)
		if code != 0 {
			t.Fatalf("pr %s: exit = %d, stderr: %s", tc.pr, code, stderr.String())
		}
		if got := strings.TrimSpace(stdout.String()); got != tc.want {
			t.Fatalf("pr %s: state = %q, want %q", tc.pr, got, tc.want)
		}
	}
}
