package cli

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/kyonRay/fisco-bcos-harness/internal/gh"
)

func init() {
	Register(Command{
		Name:    "gh",
		Summary: "query GitHub PR state (read-only; no dry-run needed)",
		Exec:    ghExec,
		// Read-only queries execute directly; no Action, no Dispatch.
	})
}

func ghExec(c *Context) error {
	if len(c.Args) == 0 {
		return fmt.Errorf("usage: fbh gh <review-state>")
	}
	sub, args := c.Args[0], c.Args[1:]
	switch sub {
	case "review-state":
		return ghReviewState(c, args)
	case "my-prs":
		return ghMyPRs(c)
	case "my-reviews":
		return ghMyReviews(c)
	default:
		return fmt.Errorf("unknown gh subcommand %q (available: review-state, my-prs, my-reviews)", sub)
	}
}

// ghMyPRs lists the caller's open PRs with their review decision and
// last review activity — the raw material for nudging and standup.
func ghMyPRs(c *Context) error {
	prs, err := gh.MyPRs()
	if err != nil {
		return err
	}
	if len(prs) == 0 {
		fmt.Fprintln(c.Stdout, "no open PRs")
		return nil
	}
	for _, pr := range prs {
		last := "never"
		if t := pr.LastReviewAt(); !t.IsZero() {
			last = t.Format("2006-01-02 15:04")
		}
		fmt.Fprintf(c.Stdout, "%s\t%s\tlast-review: %s\treviewers: %s\n",
			pr.URL, pr.NormalizedDecision(), last, strings.Join(pr.Reviewers(), ","))
	}
	return nil
}

// ghReviewState prints the ADR-0003 verdict carrier: the PR's native
// review decision, normalized to approved / changes_requested / none.
func ghReviewState(c *Context, args []string) error {
	flags, _, err := parseFlags(args)
	if err != nil {
		return err
	}
	if flags["pr"] == "" {
		return fmt.Errorf("usage: fbh gh review-state --pr <PR链接或编号>")
	}
	out, err := gh.Run("pr", "view", flags["pr"], "--json", "reviewDecision")
	if err != nil {
		return err
	}
	var parsed struct {
		ReviewDecision string `json:"reviewDecision"`
	}
	if err := json.Unmarshal([]byte(out), &parsed); err != nil {
		return fmt.Errorf("parse gh output: %w", err)
	}
	switch parsed.ReviewDecision {
	case "APPROVED":
		fmt.Fprintln(c.Stdout, "approved")
	case "CHANGES_REQUESTED":
		fmt.Fprintln(c.Stdout, "changes_requested")
	default:
		fmt.Fprintln(c.Stdout, "none")
	}
	return nil
}

// ghMyReviews lists open PRs where I owe review work, with the
// four-way loop state (approved-by-me tasks are filtered out).
func ghMyReviews(c *Context) error {
	tasks, err := gh.MyReviewTasks()
	if err != nil {
		return err
	}
	me, err := gh.Login()
	if err != nil {
		return err
	}
	owed := 0
	for _, pr := range tasks {
		state := loopState(pr, me)
		if state == "" {
			continue
		}
		owed++
		fmt.Fprintf(c.Stdout, "%s\t%s\n", pr.URL, state)
	}
	if owed == 0 {
		fmt.Fprintln(c.Stdout, "no reviews owed")
	}
	return nil
}
