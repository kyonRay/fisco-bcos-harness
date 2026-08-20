package cli

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/kyonRay/fisco-bcos-harness/internal/action"
	"github.com/kyonRay/fisco-bcos-harness/internal/schema"
)

func init() {
	Register(Command{
		Name:     "gate",
		Summary:  "run the milestone gate (wraps the release-gate command), write results back",
		Exec:     gateExec,
		Dispatch: routeAction,
	})
}

func gateExec(c *Context) error {
	if len(c.Args) == 0 || c.Args[0] != "run" {
		return fmt.Errorf("usage: fbh gate run --milestone <名> --milestone-table <id> --reqs-table <id> --owner <负责人> [--assignee <归因作者>]")
	}
	flags, _, err := parseFlags(c.Args[1:])
	if err != nil {
		return err
	}
	for _, req := range []string{"milestone", "milestone-table", "reqs-table", "owner"} {
		if flags[req] == "" {
			return fmt.Errorf("--%s is required", req)
		}
	}
	cfg, err := requireSheetConfig()
	if err != nil {
		return err
	}
	if cfg.GateCmd == "" {
		return fmt.Errorf("gate_cmd is not configured; set it with: fbh config set gate_cmd '<release-gate 调用命令>'")
	}

	// Run the gate itself (local execution; the external side effects
	// below still honor --dry-run).
	fmt.Fprintf(c.Stdout, "running gate: %s\n", cfg.GateCmd)
	var out bytes.Buffer
	// gate_cmd is the operator's own command from their 0600-perm local
	// config (like an npm script) — running it through a shell is the
	// feature, not an injection surface.
	gate := exec.Command("bash", "-c", cfg.GateCmd)
	gate.Stdout = &out
	gate.Stderr = &out
	gateErr := gate.Run()
	tail := outputTail(out.String(), 800)

	milestone := flags["milestone"]
	if gateErr == nil {
		// Pass: stamp the milestone row, complete its requirements.
		if err := c.Do(upsertAction(cfg.SheetFileID, flags["milestone-table"],
			schema.ColMilestoneName, milestone, map[string]any{
				schema.ColMilestoneOwner: flags["owner"],
				schema.ColGateStatus:     "通过",
				schema.ColGateTime:       time.Now().Format("2006-01-02 15:04"),
			})); err != nil {
			return err
		}
		if err := c.Do(action.Action{
			Service: "sheet",
			Op:      "complete_milestone",
			Payload: map[string]any{
				"file_id":   cfg.SheetFileID,
				"sheet_id":  flags["reqs-table"],
				"milestone": milestone,
			},
		}); err != nil {
			return err
		}
		fmt.Fprintf(c.Stdout, "gate passed: milestone %s completed\n", milestone)
		return nil
	}

	// Fail: file a defect row and nudge the attributed author (the
	// /fbh-gate skill does the semantic attribution; fallback = owner).
	assignee := flags["assignee"]
	if assignee == "" {
		assignee = flags["owner"]
	}
	defectKey := fmt.Sprintf("gate-defect-%s-%s", milestone, time.Now().Format("0102-1504"))
	if err := c.Do(upsertAction(cfg.SheetFileID, flags["reqs-table"],
		schema.ColRequirement, defectKey, map[string]any{
			schema.ColMilestone: milestone,
			schema.ColOwner:     assignee,
			schema.ColStatus:    "待认领",
			schema.ColNotes:     "gate-defect: " + tail,
		})); err != nil {
		return err
	}
	if err := c.Do(nudgeAction(assignee, defectKey,
		fmt.Sprintf("milestone %s 门禁失败，缺陷行已建，请认领排查。失败尾部输出：%s", milestone, tail))); err != nil {
		return err
	}
	return fmt.Errorf("gate failed for milestone %s (defect row %s filed)", milestone, defectKey)
}

func outputTail(s string, max int) string {
	s = strings.TrimSpace(s)
	runes := []rune(s) // cut on rune boundaries: gate output is Chinese-heavy
	if len(runes) <= max {
		return s
	}
	return "…" + string(runes[len(runes)-max:])
}
