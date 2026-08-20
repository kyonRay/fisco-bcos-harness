package cli

import (
	"encoding/json"
	"fmt"

	"github.com/kyonRay/fisco-bcos-harness/internal/action"
	"github.com/kyonRay/fisco-bcos-harness/internal/config"
	"github.com/kyonRay/fisco-bcos-harness/internal/mcp"
)

// sheetClient builds an MCP client from local config + the mcporter
// token saved by the tencent-docs auth flow.
func sheetClient() (*mcp.Client, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}
	token, err := mcp.LoadToken()
	if err != nil {
		return nil, err
	}
	base := cfg.MCPBaseURL
	if base == "" {
		base = mcp.DefaultBaseURL
	}
	return &mcp.Client{BaseURL: base, Token: token}, nil
}

// sheetToolName maps fbh sheet action ops to Tencent smartsheet MCP tools.
var sheetToolName = map[string]string{
	"list_tables":    "smartsheet.list_tables",
	"list_fields":    "smartsheet.list_fields",
	"list_records":   "smartsheet.list_records",
	"add_records":    "smartsheet.add_records",
	"update_records": "smartsheet.update_records",
}

func sheetDispatch(c *Context, a action.Action) error {
	tool, ok := sheetToolName[a.Op]
	if !ok {
		return fmt.Errorf("no smartsheet tool for op %q", a.Op)
	}
	client, err := sheetClient()
	if err != nil {
		return err
	}
	text, err := client.CallTool(tool, a.Payload)
	if err != nil {
		return err
	}
	return renderSheetResult(c, a.Op, text)
}

// renderSheetResult prints a human summary where the shape is known,
// falling back to the raw tool text.
func renderSheetResult(c *Context, op, text string) error {
	if op == "list_tables" {
		var parsed struct {
			Sheets []struct {
				SheetID string `json:"sheet_id"`
				Name    string `json:"name"`
			} `json:"sheets"`
		}
		if err := json.Unmarshal([]byte(text), &parsed); err == nil && len(parsed.Sheets) > 0 {
			for _, s := range parsed.Sheets {
				fmt.Fprintf(c.Stdout, "%s\t%s\n", s.SheetID, s.Name)
			}
			return nil
		}
	}
	fmt.Fprintln(c.Stdout, text)
	return nil
}
