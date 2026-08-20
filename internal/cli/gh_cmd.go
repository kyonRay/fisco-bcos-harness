package cli

import (
	"encoding/json"
	"fmt"

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
	default:
		return fmt.Errorf("unknown gh subcommand %q (available: review-state)", sub)
	}
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
