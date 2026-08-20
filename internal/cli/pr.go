package cli

import (
	"fmt"

	"github.com/kyonRay/fisco-bcos-harness/internal/action"
	"github.com/kyonRay/fisco-bcos-harness/internal/gh"
	"github.com/kyonRay/fisco-bcos-harness/internal/schema"
)

func init() {
	Register(Command{
		Name:     "pr",
		Summary:  "open a PR and run the full chain: create → sheet write-back → WeCom nudge",
		Exec:     prExec,
		Dispatch: prDispatch,
	})
}

func prExec(c *Context) error {
	if len(c.Args) == 0 || c.Args[0] != "open" {
		return fmt.Errorf("usage: fbh pr open --title <t> --body <b> --reviewer <r> --table <sheet_id> --key <需求名>")
	}
	flags, _, err := parseFlags(c.Args[1:])
	if err != nil {
		return err
	}
	for _, req := range []string{"title", "reviewer", "table", "key"} {
		if flags[req] == "" {
			return fmt.Errorf("--%s is required (usage: fbh pr open --title <t> --body <b> --reviewer <r> --table <sheet_id> --key <需求名>)", req)
		}
	}
	cfg, err := requireSheetConfig()
	if err != nil {
		return err
	}

	// 1. Create the PR with the reviewer assigned.
	if err := c.Do(action.Action{
		Service: "gh",
		Op:      "create_pr",
		Payload: map[string]any{
			"title":    flags["title"],
			"body":     flags["body"],
			"reviewer": flags["reviewer"],
		},
	}); err != nil {
		return err
	}
	prURL := c.LastResult

	// 2. Write the PR link back to the requirement row (embedded sync).
	if err := c.Do(upsertAction(cfg.SheetFileID, flags["table"], schema.ColRequirement, flags["key"],
		map[string]any{
			schema.ColPRLink: prURL,
			schema.ColStatus: "待review",
		})); err != nil {
		return err
	}

	// 3. Directed WeCom mention to the assigned reviewer.
	return c.Do(nudgeAction(flags["reviewer"], prURL, "新 PR 请 review"))
}

// prDispatch routes the chain's actions to the owning dispatcher.
func prDispatch(c *Context, a action.Action) error {
	switch a.Service {
	case "gh":
		if a.Op != "create_pr" {
			return fmt.Errorf("unknown gh op %q", a.Op)
		}
		args := []string{"pr", "create",
			"--title", payloadStr(a.Payload, "title"),
			"--body", payloadStr(a.Payload, "body"),
			"--reviewer", payloadStr(a.Payload, "reviewer"),
		}
		url, err := gh.Run(args...)
		if err != nil {
			return err
		}
		c.LastResult = url
		fmt.Fprintf(c.Stdout, "created %s\n", url)
		return nil
	default:
		return routeAction(c, a)
	}
}
