package cli

import (
	"fmt"

	"github.com/kyonRay/fisco-bcos-harness/internal/config"
)

func init() {
	Register(Command{
		Name:    "config",
		Summary: "set or show local fbh configuration (credentials stay on this machine)",
		Exec:    configExec,
		// config touches only the local file; no external Action, no Dispatch.
	})
}

func configExec(c *Context) error {
	if len(c.Args) == 0 {
		return fmt.Errorf("usage: fbh config <set|show>")
	}
	switch c.Args[0] {
	case "set":
		if len(c.Args) != 3 {
			return fmt.Errorf("usage: fbh config set <sheet_file_id|wecom_webhook|mcp_base_url> <value>")
		}
		return configSet(c, c.Args[1], c.Args[2])
	case "show":
		return configShow(c)
	default:
		return fmt.Errorf("unknown config subcommand %q (available: set, show)", c.Args[0])
	}
}

func configSet(c *Context, key, value string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	switch key {
	case "sheet_file_id":
		cfg.SheetFileID = value
	case "wecom_webhook":
		cfg.WecomWebhook = value
	case "mcp_base_url":
		cfg.MCPBaseURL = value
	default:
		return fmt.Errorf("unknown config key %q (available: sheet_file_id, wecom_webhook, mcp_base_url)", key)
	}
	if err := config.Save(cfg); err != nil {
		return err
	}
	fmt.Fprintf(c.Stdout, "set %s\n", key)
	return nil
}

func configShow(c *Context) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	fmt.Fprintf(c.Stdout, "config file:   %s\n", config.Path())
	fmt.Fprintf(c.Stdout, "sheet_file_id: %s\n", orUnset(cfg.SheetFileID))
	fmt.Fprintf(c.Stdout, "wecom_webhook: %s\n", redactWebhook(cfg.WecomWebhook))
	fmt.Fprintf(c.Stdout, "mcp_base_url:  %s\n", orUnset(cfg.MCPBaseURL))
	return nil
}

func orUnset(v string) string {
	if v == "" {
		return "(unset)"
	}
	return v
}

// redactWebhook keeps only the host part of the webhook URL so `config
// show` never echoes the shared key.
func redactWebhook(v string) string {
	if v == "" {
		return "(unset)"
	}
	return "(set, redacted)"
}
