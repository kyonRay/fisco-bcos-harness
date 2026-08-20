package cli

import (
	"fmt"
	"time"

	"github.com/kyonRay/fisco-bcos-harness/internal/gh"
)

func init() {
	Register(Command{
		Name:    "standup",
		Summary: "morning anti-miss board: my PRs awaiting review + reviews I owe",
		Exec:    standupExec,
		// Read-only board; no external Action.
	})
}

func standupExec(c *Context) error {
	myPRs, err := gh.MyPRs()
	if err != nil {
		return err
	}
	tasks, err := gh.MyReviewTasks()
	if err != nil {
		return err
	}
	me, err := gh.Login()
	if err != nil {
		return err
	}
	nudges, err := loadNudgeState()
	if err != nil {
		return err
	}

	pending := 0
	fmt.Fprintln(c.Stdout, "## 我的待 review PR")
	for _, pr := range myPRs {
		if pr.ReviewDecision == "APPROVED" {
			continue
		}
		pending++
		line := fmt.Sprintf("- %s  %s  挂了 %s", pr.URL, pr.NormalizedDecision(), sinceHuman(pr.UpdatedAt))
		if n := nudges[pr.URL].Count; n > 0 {
			line += fmt.Sprintf("  催过 %d 次", n)
		}
		fmt.Fprintln(c.Stdout, line)
	}

	owed := 0
	fmt.Fprintln(c.Stdout, "\n## 我欠的 review")
	for _, pr := range tasks {
		state := loopState(pr, me)
		if state == "" {
			continue // approved by me: nothing owed
		}
		owed++
		fmt.Fprintf(c.Stdout, "- %s  %s\n", pr.URL, state)
	}

	if pending == 0 && owed == 0 {
		fmt.Fprintln(c.Stdout, "\n无欠账 ✅")
	}
	return nil
}

// loopState classifies one review task into the harness's four loop
// states, from my own newest review + activity since it. A COMMENTED
// review is the pre-read brief (/fbh-review-pr posts it that way), so
// it means "brief posted, human approve pending".
func loopState(pr gh.PR, me string) string {
	state, at := pr.MyLastReview(me)
	switch state {
	case "":
		return "等我首审"
	case "COMMENTED":
		return "等人工approve"
	case "CHANGES_REQUESTED":
		if pr.UpdatedAt.After(at) {
			return "等我复审"
		}
		return "等作者改"
	case "APPROVED":
		return ""
	default:
		return "等我首审"
	}
}

func sinceHuman(t time.Time) string {
	d := time.Since(t)
	if d < time.Hour {
		return fmt.Sprintf("%d 分钟", int(d.Minutes()))
	}
	if d < 48*time.Hour {
		return fmt.Sprintf("%d 小时", int(d.Hours()))
	}
	return fmt.Sprintf("%d 天", int(d.Hours()/24))
}
