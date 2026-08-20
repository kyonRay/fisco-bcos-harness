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

func nudgeEnv(t *testing.T) *fakeWecom {
	t.Helper()
	wecom := &fakeWecom{}
	srv := wecom.start(t)
	t.Cleanup(srv.Close)
	cfgPath := filepath.Join(t.TempDir(), "config.json")
	t.Setenv("FBH_CONFIG", cfgPath)
	cfg := `{"wecom_webhook":"` + srv.URL + `","sheet_file_id":"SHEET123"}`
	if err := os.WriteFile(cfgPath, []byte(cfg), 0o600); err != nil {
		t.Fatal(err)
	}
	return wecom
}

func TestNudgeRunSendsOnceThenIdempotentSameDay(t *testing.T) {
	wecom := nudgeEnv(t)
	prsFixture(t, stalePR(25))
	var stdout, stderr bytes.Buffer

	if code := Run([]string{"nudge", "run"}, &stdout, &stderr); code != 0 {
		t.Fatalf("first run exit = %d, stderr: %s", code, stderr.String())
	}
	if len(wecom.bodies) != 1 {
		t.Fatalf("first run sent %d nudges, want exactly 1", len(wecom.bodies))
	}
	if !strings.Contains(wecom.bodies[0], "lisi") || !strings.Contains(wecom.bodies[0], "pull/42") {
		t.Fatalf("nudge must @lisi with PR link, got: %s", wecom.bodies[0])
	}

	if code := Run([]string{"nudge", "run"}, &stdout, &stderr); code != 0 {
		t.Fatalf("second run exit = %d, stderr: %s", code, stderr.String())
	}
	if len(wecom.bodies) != 1 {
		t.Fatalf("second run same day must send 0 more, total = %d", len(wecom.bodies))
	}
}

func TestNudgeRunSkipsFreshPR(t *testing.T) {
	wecom := nudgeEnv(t)
	prsFixture(t, stalePR(2)) // only 2h old, under the 24h default threshold
	var stdout, stderr bytes.Buffer

	if code := Run([]string{"nudge", "run"}, &stdout, &stderr); code != 0 {
		t.Fatalf("exit = %d, stderr: %s", code, stderr.String())
	}
	if len(wecom.bodies) != 0 {
		t.Fatalf("fresh PR must not be nudged, sent %d", len(wecom.bodies))
	}
}

func TestNudgeRunSkipsChangesRequestedWaitingOnAuthor(t *testing.T) {
	wecom := nudgeEnv(t)
	prsFixture(t, changesRequestedWaitingAuthor())
	var stdout, stderr bytes.Buffer

	if code := Run([]string{"nudge", "run"}, &stdout, &stderr); code != 0 {
		t.Fatalf("exit = %d, stderr: %s", code, stderr.String())
	}
	if len(wecom.bodies) != 0 {
		t.Fatalf("waiting-on-author PR must not nudge the reviewer, sent %d: %s", len(wecom.bodies), wecom.bodies)
	}
}

func TestNudgeRunDryRunPreviewsWithoutRecordingState(t *testing.T) {
	wecom := nudgeEnv(t)
	prsFixture(t, stalePR(25))
	var stdout, stderr bytes.Buffer

	if code := Run([]string{"nudge", "run", "--dry-run"}, &stdout, &stderr); code != 0 {
		t.Fatalf("dry-run exit = %d, stderr: %s", code, stderr.String())
	}
	if len(wecom.bodies) != 0 {
		t.Fatalf("dry-run must not send, sent %d", len(wecom.bodies))
	}
	if !strings.Contains(stdout.String(), `"nudge"`) {
		t.Fatalf("dry-run must print the pending action, got: %s", stdout.String())
	}

	// A real run afterwards must still send: dry-run didn't burn the day.
	stdout.Reset()
	if code := Run([]string{"nudge", "run"}, &stdout, &stderr); code != 0 {
		t.Fatalf("real run exit = %d, stderr: %s", code, stderr.String())
	}
	if len(wecom.bodies) != 1 {
		t.Fatalf("real run after dry-run must send 1, sent %d", len(wecom.bodies))
	}
}

func TestNudgeThresholdConfigurable(t *testing.T) {
	wecom := nudgeEnv(t)
	prsFixture(t, stalePR(2))
	var stdout, stderr bytes.Buffer

	if code := Run([]string{"nudge", "run", "--threshold", "1"}, &stdout, &stderr); code != 0 {
		t.Fatalf("exit = %d, stderr: %s", code, stderr.String())
	}
	if len(wecom.bodies) != 1 {
		t.Fatalf("2h-old PR with 1h threshold must nudge, sent %d", len(wecom.bodies))
	}
}

func TestGhMyPrsListsWithLastReviewActivity(t *testing.T) {
	nudgeEnv(t)
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
