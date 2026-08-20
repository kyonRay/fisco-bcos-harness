package cli

import (
	"fmt"
	"strings"

	"github.com/kyonRay/fisco-bcos-harness/internal/action"
	"github.com/kyonRay/fisco-bcos-harness/internal/config"
	"github.com/kyonRay/fisco-bcos-harness/internal/schema"
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
		return fmt.Errorf("usage: fbh sheet <ping|upsert-row|set-status|claim>")
	}
	sub, args := c.Args[0], c.Args[1:]
	switch sub {
	case "ping":
		return sheetPing(c)
	case "upsert-row":
		return sheetUpsertRow(c, args)
	case "set-status":
		return sheetSetStatus(c, args)
	case "claim":
		return sheetClaim(c, args)
	default:
		return fmt.Errorf("unknown sheet subcommand %q (available: ping, upsert-row, set-status, claim)", sub)
	}
}

func requireSheetConfig() (config.Config, error) {
	cfg, err := config.Load()
	if err != nil {
		return cfg, err
	}
	if cfg.SheetFileID == "" {
		return cfg, fmt.Errorf("sheet_file_id is not configured; run the /fbh-setup skill first")
	}
	return cfg, nil
}

func sheetPing(c *Context) error {
	cfg, err := requireSheetConfig()
	if err != nil {
		return err
	}
	return c.Do(action.Action{
		Service: "sheet",
		Op:      "list_tables",
		Payload: map[string]any{"file_id": cfg.SheetFileID},
	})
}

// parseFlags consumes --flag value pairs plus repeated --set k=v.
func parseFlags(args []string) (flags map[string]string, sets map[string]any, err error) {
	flags = map[string]string{}
	sets = map[string]any{}
	for i := 0; i < len(args); i++ {
		if !strings.HasPrefix(args[i], "--") {
			return nil, nil, fmt.Errorf("unexpected argument %q", args[i])
		}
		name := strings.TrimPrefix(args[i], "--")
		if i+1 >= len(args) {
			return nil, nil, fmt.Errorf("--%s needs a value", name)
		}
		val := args[i+1]
		i++
		if name == "set" {
			kv := strings.SplitN(val, "=", 2)
			if len(kv) != 2 {
				return nil, nil, fmt.Errorf("--set expects 列名=值, got %q", val)
			}
			sets[kv[0]] = kv[1]
			continue
		}
		flags[name] = val
	}
	return flags, sets, nil
}

func statusError(bad string) error {
	return fmt.Errorf("invalid status %q; valid statuses: %s", bad, strings.Join(schema.Statuses, " → "))
}

func upsertAction(fileID, sheetID, keyField, key string, sets map[string]any) action.Action {
	return action.Action{
		Service: "sheet",
		Op:      "upsert_row",
		Payload: map[string]any{
			"file_id":   fileID,
			"sheet_id":  sheetID,
			"key_field": keyField,
			"key":       key,
			"sets":      sets,
		},
	}
}

func sheetUpsertRow(c *Context, args []string) error {
	cfg, err := requireSheetConfig()
	if err != nil {
		return err
	}
	flags, sets, err := parseFlags(args)
	if err != nil {
		return err
	}
	if flags["table"] == "" || flags["key"] == "" || len(sets) == 0 {
		return fmt.Errorf("usage: fbh sheet upsert-row --table <sheet_id> --key <需求名> --set 列名=值 [--set ...]")
	}
	if s, ok := sets[schema.ColStatus].(string); ok && !schema.ValidStatus(s) {
		return statusError(s)
	}
	keyField := flags["key-field"]
	if keyField == "" {
		keyField = schema.ColRequirement
	}
	return c.Do(upsertAction(cfg.SheetFileID, flags["table"], keyField, flags["key"], sets))
}

func sheetSetStatus(c *Context, args []string) error {
	cfg, err := requireSheetConfig()
	if err != nil {
		return err
	}
	flags, _, err := parseFlags(args)
	if err != nil {
		return err
	}
	if flags["table"] == "" || flags["key"] == "" || flags["status"] == "" {
		return fmt.Errorf("usage: fbh sheet set-status --table <sheet_id> --key <需求名> --status <状态>")
	}
	if !schema.ValidStatus(flags["status"]) {
		return statusError(flags["status"])
	}
	keyField := flags["key-field"]
	if keyField == "" {
		keyField = schema.ColRequirement
	}
	sets := map[string]any{schema.ColStatus: flags["status"]}
	return c.Do(upsertAction(cfg.SheetFileID, flags["table"], keyField, flags["key"], sets))
}

func sheetClaim(c *Context, args []string) error {
	cfg, err := requireSheetConfig()
	if err != nil {
		return err
	}
	flags, _, err := parseFlags(args)
	if err != nil {
		return err
	}
	if flags["table"] == "" || flags["key"] == "" || flags["owner"] == "" {
		return fmt.Errorf("usage: fbh sheet claim --table <sheet_id> --key <需求名> --owner <认领人>")
	}
	return c.Do(action.Action{
		Service: "sheet",
		Op:      "claim",
		Payload: map[string]any{
			"file_id":  cfg.SheetFileID,
			"sheet_id": flags["table"],
			"key":      flags["key"],
			"owner":    flags["owner"],
		},
	})
}
