package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/kyonRay/fisco-bcos-harness/internal/config"
	"github.com/kyonRay/fisco-bcos-harness/internal/gh"
)

func init() {
	Register(Command{
		Name:     "nudge",
		Summary:  "author-driven review nudging: detect stale own PRs, WeCom-mention their reviewers",
		Exec:     nudgeExec,
		Dispatch: wecomDispatch, // the only external action nudge emits is wecom.nudge
	})
}

func nudgeExec(c *Context) error {
	if len(c.Args) == 0 || c.Args[0] != "run" {
		return fmt.Errorf("usage: fbh nudge run [--threshold <hours>]")
	}
	flags, _, err := parseFlags(c.Args[1:])
	if err != nil {
		return err
	}
	threshold := 24 * time.Hour
	if v := flags["threshold"]; v != "" {
		hours, err := strconv.Atoi(v)
		if err != nil || hours <= 0 {
			return fmt.Errorf("--threshold expects a positive hour count, got %q", v)
		}
		threshold = time.Duration(hours) * time.Hour
	}

	prs, err := gh.MyPRs()
	if err != nil {
		return err
	}
	state, err := loadNudgeState()
	if err != nil {
		return err
	}
	today := time.Now().Format("2006-01-02")
	nudged := 0
	for _, pr := range prs {
		if !shouldNudge(pr, threshold) {
			continue
		}
		if state[pr.URL].Date == today {
			continue // already nudged today: idempotent
		}
		for _, reviewer := range pr.Reviewers() {
			if err := c.Do(nudgeAction(reviewer, pr.URL, "PR 等 review 超时，请抽空看看")); err != nil {
				return err
			}
		}
		state[pr.URL] = nudgeRecord{Date: today, Count: state[pr.URL].Count + 1}
		nudged++
		// Save per PR (not once at the end): if a later send fails,
		// already-sent nudges stay recorded and won't repeat on rerun.
		// Dry-run previews must not burn today's budget.
		if !c.dryRun {
			if err := saveNudgeState(state); err != nil {
				return err
			}
		}
	}
	if nudged == 0 {
		fmt.Fprintln(c.Stdout, "nothing to nudge")
	}
	return nil
}

// shouldNudge implements the staleness rule:
//   - approved PRs never nudge;
//   - changes_requested with no activity since the review means the
//     ball is in the author's court — nudging the reviewer would be noise;
//   - otherwise nudge when the last activity (updatedAt) is older than
//     the threshold.
func shouldNudge(pr gh.PR, threshold time.Duration) bool {
	switch pr.ReviewDecision {
	case "APPROVED":
		return false
	case "CHANGES_REQUESTED":
		if !pr.UpdatedAt.After(pr.LastReviewAt()) {
			return false
		}
	}
	if len(pr.Reviewers()) == 0 {
		return false
	}
	return time.Since(pr.UpdatedAt) >= threshold
}

// nudgeRecord tracks per-PR nudge bookkeeping: last nudge day (for
// same-day idempotency) and lifetime count (shown by standup).
type nudgeRecord struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

func nudgeStatePath() string {
	return filepath.Join(filepath.Dir(config.Path()), "nudge-state.json")
}

func loadNudgeState() (map[string]nudgeRecord, error) {
	state := map[string]nudgeRecord{}
	data, err := os.ReadFile(nudgeStatePath())
	if os.IsNotExist(err) {
		return state, nil
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("parse %s: %w", nudgeStatePath(), err)
	}
	return state, nil
}

func saveNudgeState(state map[string]nudgeRecord) error {
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(nudgeStatePath()), 0o700); err != nil {
		return err
	}
	return os.WriteFile(nudgeStatePath(), append(data, '\n'), 0o600)
}
