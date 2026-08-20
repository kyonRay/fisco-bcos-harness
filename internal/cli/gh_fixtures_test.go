package cli

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// prsFixture builds the gh pr list JSON the stub echoes back.
func prsFixture(t *testing.T, entries ...string) {
	t.Helper()
	t.Setenv("FBH_TEST_PRS", "["+strings.Join(entries, ",")+"]")
	writeGhStub(t, `case "$1 $2" in
"pr list") echo "$FBH_TEST_PRS" ;;
*) echo "unexpected args: $*" >&2; exit 1 ;;
esac`)
}

func ago(h int) string { return time.Now().Add(-time.Duration(h) * time.Hour).Format(time.RFC3339) }

func stalePR(hoursOld int) string {
	return fmt.Sprintf(`{"number":42,"url":"https://github.com/team/repo/pull/42","title":"feat: login",
"createdAt":%q,"updatedAt":%q,"reviewDecision":"",
"reviewRequests":[{"login":"lisi"}],"latestReviews":[]}`, ago(hoursOld), ago(hoursOld))
}

func changesRequestedWaitingAuthor() string {
	return fmt.Sprintf(`{"number":43,"url":"https://github.com/team/repo/pull/43","title":"fix: x",
"createdAt":%q,"updatedAt":%q,"reviewDecision":"CHANGES_REQUESTED",
"reviewRequests":[],"latestReviews":[{"author":{"login":"lisi"},"state":"CHANGES_REQUESTED","submittedAt":%q}]}`,
		ago(50), ago(30), ago(30))
}

// ghEnv isolates the config file so tests never touch ~/.fbh.
func ghEnv(t *testing.T) {
	t.Helper()
	cfgPath := filepath.Join(t.TempDir(), "config.json")
	t.Setenv("FBH_CONFIG", cfgPath)
	if err := os.WriteFile(cfgPath, []byte(`{"sheet_file_id":"SHEET123"}`), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestGhMyPrsListsWithLastReviewActivity(t *testing.T) {
	ghEnv(t)
	prsFixture(t, changesRequestedWaitingAuthor())
	var stdout, stderr bytes.Buffer

	if code := Run([]string{"gh", "my-prs"}, &stdout, &stderr); code != 0 {
		t.Fatalf("exit = %d, stderr: %s", code, stderr.String())
	}
	out := stdout.String()
	for _, want := range []string{"pull/43", "changes_requested", "lisi"} {
		if !strings.Contains(out, want) {
			t.Fatalf("my-prs output missing %q, got: %s", want, out)
		}
	}
}
