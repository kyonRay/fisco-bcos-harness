package cli

import (
	"fmt"
	"strings"
	"time"

	"github.com/kyonRay/fisco-bcos-harness/internal/action"
	"github.com/kyonRay/fisco-bcos-harness/internal/config"
	"github.com/kyonRay/fisco-bcos-harness/internal/gh"
	"github.com/kyonRay/fisco-bcos-harness/internal/schema"
)

func init() {
	Register(Command{
		Name:     "pr",
		Summary:  "open a PR and register it in the ledger; sheet automation sends the notifications",
		Exec:     prExec,
		Dispatch: prDispatch,
	})
}

func prExec(c *Context) error {
	if len(c.Args) == 0 {
		return fmt.Errorf("usage: fbh pr <open|sync|board>")
	}
	sub, rest := c.Args[0], c.Args[1:]
	switch sub {
	case "open":
		return prOpen(c, rest)
	case "sync":
		return prSync(c, rest)
	case "board":
		return prBoard(c)
	default:
		return fmt.Errorf("unknown pr subcommand %q (available: open, sync, board)", sub)
	}
}

func prOpen(c *Context, args []string) error {
	flags, _, err := parseFlags(args)
	if err != nil {
		return err
	}
	for _, req := range []string{"title", "reviewer"} {
		if flags[req] == "" {
			return fmt.Errorf("--%s is required (usage: fbh pr open --title <t> --body <b> --reviewer <r> [--table <sheet_id> --key <需求名>])", req)
		}
	}
	// The sheet write-back is optional: minimal adoption (只治 review)
	// runs without any Tencent-sheet configuration. It engages only
	// when --table/--key are passed, which then requires sheet config.
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	writeBack := flags["table"] != "" || flags["key"] != ""
	if writeBack {
		if flags["table"] == "" || flags["key"] == "" {
			return fmt.Errorf("--table and --key must be given together")
		}
		if cfg.SheetFileID == "" {
			return fmt.Errorf("--table given but sheet_file_id is not configured; run the /fbh-setup skill first")
		}
	}
	if cfg.PRSheetID != "" && cfg.SheetFileID == "" {
		return fmt.Errorf("pr_sheet_id is set but sheet_file_id is not; run the /fbh-setup skill first")
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

	// 2a. Register the PR in the PR ledger (PR 台账模式): rows keyed by
	// PR URL, so the whole team's AIs can see what awaits review. The
	// row add itself is the notification: the sheet's 添加记录提醒到群
	// automation announces it — fbh sends nothing directly.
	if cfg.PRSheetID != "" {
		author, err := gh.Login()
		if err != nil {
			return err
		}
		if err := c.Do(upsertAction(cfg.SheetFileID, cfg.PRSheetID, schema.ColPRURL, prURL,
			map[string]any{
				schema.ColPRTitle:     flags["title"],
				schema.ColPRAuthor:    author,
				schema.ColPRReviewers: flags["reviewer"],
				schema.ColPRApproved:  "",
				schema.ColPRStatus:    "待review",
				schema.ColPRPending:   flags["reviewer"],
				schema.ColPRUpdated:   time.Now().Format("2006-01-02 15:04"),
			})); err != nil {
			return err
		}
	}

	// 2b. Write the PR link back to the requirement row (embedded
	// sync), when the requirement half of the workflow is in use.
	if writeBack {
		if err := c.Do(upsertAction(cfg.SheetFileID, flags["table"], schema.ColRequirement, flags["key"],
			map[string]any{
				schema.ColPRLink: prURL,
				schema.ColStatus: "待review",
			})); err != nil {
			return err
		}
	}
	return nil
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

// prLedgerStatus derives the PR-ledger status from GitHub state.
func prLedgerStatus(pr gh.PR) string {
	if pr.State == "MERGED" {
		return "已合入"
	}
	if pr.ReviewDecision == "APPROVED" {
		return "已approve"
	}
	if pr.ReviewDecision == "CHANGES_REQUESTED" {
		if pr.UpdatedAt.After(pr.LastReviewAt()) {
			return "待复审" // author pushed fixes after the review
		}
		return "修复中"
	}
	return "待review"
}

// prSync refreshes one PR's ledger row from GitHub. 待处理人 becomes
// the reviewers who have not approved yet; the sheet's 修改后提醒负责人
// automation notifies them off that column — fbh sends nothing itself.
func prSync(c *Context, args []string) error {
	flags, _, err := parseFlags(args)
	if err != nil {
		return err
	}
	if flags["pr"] == "" {
		return fmt.Errorf("usage: fbh pr sync --pr <PR链接>")
	}
	cfg, err := requireSheetConfig()
	if err != nil {
		return err
	}
	if cfg.PRSheetID == "" {
		return fmt.Errorf("pr_sheet_id is not configured; run the /fbh-setup skill first")
	}

	pr, err := gh.ViewPR(flags["pr"])
	if err != nil {
		return err
	}
	status := prLedgerStatus(pr)
	pending := pr.PendingReviewers()
	if status == "已approve" || status == "已合入" {
		pending = nil // finished work must not keep pinging anyone
	}
	if err := c.Do(upsertAction(cfg.SheetFileID, cfg.PRSheetID, schema.ColPRURL, pr.URL,
		map[string]any{
			schema.ColPRTitle:    pr.Title,
			schema.ColPRAuthor:   pr.Author.Login,
			schema.ColPRApproved: strings.Join(pr.ApprovedBy(), ","),
			schema.ColPRStatus:   status,
			schema.ColPRPending:  strings.Join(pending, ","),
			schema.ColPRUpdated:  time.Now().Format("2006-01-02 15:04"),
		})); err != nil {
		return err
	}
	fmt.Fprintf(c.Stdout, "synced %s -> %s (待处理人: %s)\n", pr.URL, status, strings.Join(pending, ","))
	return nil
}

func prBoard(c *Context) error {
	cfg, err := requireSheetConfig()
	if err != nil {
		return err
	}
	if cfg.PRSheetID == "" {
		return fmt.Errorf("pr_sheet_id is not configured; run the /fbh-setup skill first")
	}
	client, err := sheetClient()
	if err != nil {
		return err
	}
	records, err := listAllRecords(client, cfg.SheetFileID, cfg.PRSheetID)
	if err != nil {
		return err
	}
	open := 0
	for _, rec := range records {
		status := rec.text(schema.ColPRStatus)
		if status == "已合入" || status == "已approve" {
			continue
		}
		open++
		fmt.Fprintf(c.Stdout, "%s\t%s\treviewers: %s\t已approve: %s\n",
			rec.text(schema.ColPRURL), status,
			rec.text(schema.ColPRReviewers), rec.text(schema.ColPRApproved))
	}
	if open == 0 {
		fmt.Fprintln(c.Stdout, "台账中无待 review 的 PR ✅")
	}
	return nil
}
