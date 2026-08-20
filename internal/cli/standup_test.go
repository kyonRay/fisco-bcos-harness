package cli

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

// standupStub wires gh: my author'd PRs, review-requested, reviewed-by,
// and the current login.
func standupStub(t *testing.T, myPRs, requested, reviewed string) {
	t.Helper()
	t.Setenv("FBH_TEST_MYPRS", myPRs)
	t.Setenv("FBH_TEST_REQUESTED", requested)
	t.Setenv("FBH_TEST_REVIEWED", reviewed)
	writeGhStub(t, `case "$*" in
"api user"*) echo "zhangsan" ;;
*"--author @me"*) echo "$FBH_TEST_MYPRS" ;;
*review-requested*) echo "$FBH_TEST_REQUESTED" ;;
*reviewed-by*) echo "$FBH_TEST_REVIEWED" ;;
*) echo "unexpected args: $*" >&2; exit 1 ;;
esac`)
}

// Review tasks in all four loop states.
func reviewFixtures() (requested, reviewed string) {
	// 51: nobody reviewed yet, I'm requested → 等我首审
	first := fmt.Sprintf(`{"number":51,"url":"https://github.com/t/r/pull/51","title":"a",
"createdAt":%q,"updatedAt":%q,"reviewDecision":"REVIEW_REQUIRED",
"reviewRequests":[{"login":"zhangsan"}],"latestReviews":[]}`, ago(30), ago(30))
	// 52: I request-changed, author silent since → 等作者改
	waitAuthor := fmt.Sprintf(`{"number":52,"url":"https://github.com/t/r/pull/52","title":"b",
"createdAt":%q,"updatedAt":%q,"reviewDecision":"CHANGES_REQUESTED",
"reviewRequests":[],"latestReviews":[{"author":{"login":"zhangsan"},"state":"CHANGES_REQUESTED","submittedAt":%q}]}`,
		ago(50), ago(20), ago(20))
	// 53: I request-changed, author pushed after → 等我复审
	reReview := fmt.Sprintf(`{"number":53,"url":"https://github.com/t/r/pull/53","title":"c",
"createdAt":%q,"updatedAt":%q,"reviewDecision":"CHANGES_REQUESTED",
"reviewRequests":[],"latestReviews":[{"author":{"login":"zhangsan"},"state":"CHANGES_REQUESTED","submittedAt":%q}]}`,
		ago(50), ago(5), ago(20))
	// 54: I posted the brief (COMMENTED) → 等人工approve
	brief := fmt.Sprintf(`{"number":54,"url":"https://github.com/t/r/pull/54","title":"d",
"createdAt":%q,"updatedAt":%q,"reviewDecision":"REVIEW_REQUIRED",
"reviewRequests":[],"latestReviews":[{"author":{"login":"zhangsan"},"state":"COMMENTED","submittedAt":%q}]}`,
		ago(50), ago(6), ago(6))
	return "[" + first + "]", "[" + waitAuthor + "," + reReview + "," + brief + "]"
}

func TestStandupShowsBothSectionsWithLoopStates(t *testing.T) {
	nudgeEnv(t)
	requested, reviewed := reviewFixtures()
	standupStub(t, "["+stalePR(30)+"]", requested, reviewed)
	var stdout, stderr bytes.Buffer

	if code := Run([]string{"standup"}, &stdout, &stderr); code != 0 {
		t.Fatalf("exit = %d, stderr: %s", code, stderr.String())
	}
	out := stdout.String()
	for _, want := range []string{
		"我的待 review PR", "pull/42",
		"我欠的 review",
		"pull/51", "等我首审",
		"pull/52", "等作者改",
		"pull/53", "等我复审",
		"pull/54", "等人工approve",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("standup missing %q, got:\n%s", want, out)
		}
	}
}

func TestStandupEmptySaysNoDebt(t *testing.T) {
	nudgeEnv(t)
	standupStub(t, "[]", "[]", "[]")
	var stdout, stderr bytes.Buffer

	if code := Run([]string{"standup"}, &stdout, &stderr); code != 0 {
		t.Fatalf("exit = %d, stderr: %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "无欠账") {
		t.Fatalf("empty standup must say 无欠账, got: %s", stdout.String())
	}
}

func TestStandupShowsNudgeCount(t *testing.T) {
	wecom := nudgeEnv(t)
	requested, reviewed := reviewFixtures()
	standupStub(t, "["+stalePR(30)+"]", requested, reviewed)

	// One real nudge first so the count is 1.
	var buf bytes.Buffer
	if code := Run([]string{"nudge", "run"}, &buf, &buf); code != 0 {
		t.Fatalf("nudge exit != 0: %s", buf.String())
	}
	if len(wecom.bodies) != 1 {
		t.Fatalf("expected 1 nudge sent, got %d", len(wecom.bodies))
	}

	var stdout, stderr bytes.Buffer
	if code := Run([]string{"standup"}, &stdout, &stderr); code != 0 {
		t.Fatalf("standup exit = %d, stderr: %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "催过 1 次") {
		t.Fatalf("standup must show nudge count, got:\n%s", stdout.String())
	}
}
