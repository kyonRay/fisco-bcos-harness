package cli

import (
	"fmt"

	"github.com/kyonRay/fisco-bcos-harness/internal/action"
	"github.com/kyonRay/fisco-bcos-harness/internal/config"
)

func init() {
	Register(Command{
		Name:     "sheet",
		Summary:  "operate the team's Tencent smart sheet (source of truth)",
		Exec:     sheetExec,
		Dispatch: sheetDispatch,
	})
}

func sheetExec(c *Context) error {
	if len(c.Args) == 0 {
		return fmt.Errorf("usage: fbh sheet <ping>")
	}
	switch c.Args[0] {
	case "ping":
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		if cfg.SheetFileID == "" {
			return fmt.Errorf("sheet_file_id is not configured; run the /fbh-setup skill first")
		}
		return c.Do(action.Action{
			Service: "sheet",
			Op:      "list_tables",
			Payload: map[string]any{"file_id": cfg.SheetFileID},
		})
	default:
		return fmt.Errorf("unknown sheet subcommand %q (available: ping)", c.Args[0])
	}
}
