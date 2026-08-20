package cli

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/kyonRay/fisco-bcos-harness/internal/action"
	"github.com/kyonRay/fisco-bcos-harness/internal/config"
	"github.com/kyonRay/fisco-bcos-harness/internal/mcp"
	"github.com/kyonRay/fisco-bcos-harness/internal/schema"
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
	client, err := sheetClient()
	if err != nil {
		return err
	}
	switch a.Op {
	case "upsert_row":
		return dispatchUpsertRow(c, client, a)
	case "claim":
		return dispatchClaim(c, client, a)
	case "complete_milestone":
		return dispatchCompleteMilestone(c, client, a)
	}
	tool, ok := sheetToolName[a.Op]
	if !ok {
		return fmt.Errorf("no smartsheet tool for op %q", a.Op)
	}
	text, err := client.CallTool(tool, a.Payload)
	if err != nil {
		return err
	}
	return renderSheetResult(c, a.Op, text)
}

// sheetRecord is one row as returned by smartsheet.list_records.
type sheetRecord struct {
	RecordID    string `json:"record_id"`
	FieldValues []struct {
		Field     string `json:"field"`
		TextValue struct {
			Items []struct {
				Text string `json:"text"`
			} `json:"items"`
		} `json:"text_value"`
	} `json:"field_values"`
}

func (r sheetRecord) text(field string) string {
	for _, fv := range r.FieldValues {
		if fv.Field != field {
			continue
		}
		var parts []string
		for _, item := range fv.TextValue.Items {
			parts = append(parts, item.Text)
		}
		return strings.Join(parts, "")
	}
	return ""
}

// listAllRecords fetches every row of the sheet, following
// pagination (list_records caps a page at 100 rows — without the loop,
// rows beyond page 1 would be invisible and upserts would duplicate).
func listAllRecords(client *mcp.Client, fileID, sheetID string) ([]sheetRecord, error) {
	var all []sheetRecord
	offset := 0
	for {
		text, err := client.CallTool("smartsheet.list_records", map[string]any{
			"file_id":  fileID,
			"sheet_id": sheetID,
			"offset":   offset,
			"limit":    100,
		})
		if err != nil {
			return nil, err
		}
		var parsed struct {
			Records []sheetRecord `json:"records"`
			HasMore bool          `json:"has_more"`
			Next    int           `json:"next"`
		}
		if err := json.Unmarshal([]byte(text), &parsed); err != nil {
			return nil, fmt.Errorf("parse list_records result: %w", err)
		}
		all = append(all, parsed.Records...)
		if !parsed.HasMore {
			return all, nil
		}
		offset = parsed.Next
	}
}

// findRecord looks up the row whose keyField text equals key.
func findRecord(client *mcp.Client, fileID, sheetID, keyField, key string) (*sheetRecord, error) {
	records, err := listAllRecords(client, fileID, sheetID)
	if err != nil {
		return nil, err
	}
	for i := range records {
		if records[i].text(keyField) == key {
			return &records[i], nil
		}
	}
	return nil, nil
}

// textEntries converts 列名→值 sets into smartsheet FieldValueEntry
// objects (all text-typed; the team sheet uses text columns).
func textEntries(sets map[string]any) []map[string]any {
	titles := make([]string, 0, len(sets))
	for k := range sets {
		titles = append(titles, k)
	}
	sort.Strings(titles)
	entries := make([]map[string]any, 0, len(sets))
	for _, k := range titles {
		entries = append(entries, map[string]any{
			"field": k,
			"text_value": map[string]any{
				"items": []map[string]any{{"type": "text", "text": fmt.Sprintf("%v", sets[k])}},
			},
		})
	}
	return entries
}

func payloadStr(p map[string]any, key string) string {
	v, _ := p[key].(string)
	return v
}

func payloadSets(p map[string]any) map[string]any {
	if m, ok := p["sets"].(map[string]any); ok {
		return m
	}
	return map[string]any{}
}

func dispatchUpsertRow(c *Context, client *mcp.Client, a action.Action) error {
	fileID := payloadStr(a.Payload, "file_id")
	sheetID := payloadStr(a.Payload, "sheet_id")
	keyField := payloadStr(a.Payload, "key_field")
	key := payloadStr(a.Payload, "key")
	sets := payloadSets(a.Payload)

	rec, err := findRecord(client, fileID, sheetID, keyField, key)
	if err != nil {
		return err
	}
	if rec == nil {
		full := map[string]any{keyField: key}
		for k, v := range sets {
			full[k] = v
		}
		_, err := client.CallTool("smartsheet.add_records", map[string]any{
			"file_id":  fileID,
			"sheet_id": sheetID,
			"records":  []map[string]any{{"field_values": textEntries(full)}},
		})
		if err != nil {
			return err
		}
		fmt.Fprintf(c.Stdout, "added row %q\n", key)
		return nil
	}
	_, err = client.CallTool("smartsheet.update_records", map[string]any{
		"file_id":  fileID,
		"sheet_id": sheetID,
		"records": []map[string]any{{
			"record_id":    rec.RecordID,
			"field_values": textEntries(sets),
		}},
	})
	if err != nil {
		return err
	}
	fmt.Fprintf(c.Stdout, "updated row %q\n", key)
	return nil
}

func dispatchClaim(c *Context, client *mcp.Client, a action.Action) error {
	fileID := payloadStr(a.Payload, "file_id")
	sheetID := payloadStr(a.Payload, "sheet_id")
	key := payloadStr(a.Payload, "key")
	owner := payloadStr(a.Payload, "owner")

	rec, err := findRecord(client, fileID, sheetID, schema.ColRequirement, key)
	if err != nil {
		return err
	}
	if rec == nil {
		return fmt.Errorf("requirement %q not found in sheet %s", key, sheetID)
	}
	if current := rec.text(schema.ColOwner); current != "" {
		if current == owner {
			// Re-claim by the same owner is a no-op: writing again
			// would drag an advanced status back to 开发中.
			fmt.Fprintf(c.Stdout, "already claimed by you (%s); no change\n", owner)
			return nil
		}
		return fmt.Errorf("requirement %q is already claimed by %s", key, current)
	}
	_, err = client.CallTool("smartsheet.update_records", map[string]any{
		"file_id":  fileID,
		"sheet_id": sheetID,
		"records": []map[string]any{{
			"record_id": rec.RecordID,
			"field_values": textEntries(map[string]any{
				schema.ColOwner:  owner,
				schema.ColStatus: "开发中",
			}),
		}},
	})
	if err != nil {
		return err
	}
	fmt.Fprintf(c.Stdout, "claimed %q as %s\n", key, owner)
	return nil
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

// dispatchCompleteMilestone sets 状态=完成 on every requirement row of
// the given milestone, leaving other milestones' rows untouched.
func dispatchCompleteMilestone(c *Context, client *mcp.Client, a action.Action) error {
	fileID := payloadStr(a.Payload, "file_id")
	sheetID := payloadStr(a.Payload, "sheet_id")
	milestone := payloadStr(a.Payload, "milestone")

	records, err := listAllRecords(client, fileID, sheetID)
	if err != nil {
		return err
	}
	completed := 0
	for _, rec := range records {
		if rec.text(schema.ColMilestone) != milestone {
			continue
		}
		_, err := client.CallTool("smartsheet.update_records", map[string]any{
			"file_id":  fileID,
			"sheet_id": sheetID,
			"records": []map[string]any{{
				"record_id":    rec.RecordID,
				"field_values": textEntries(map[string]any{schema.ColStatus: "完成"}),
			}},
		})
		if err != nil {
			return err
		}
		completed++
	}
	fmt.Fprintf(c.Stdout, "completed %d requirement(s) of milestone %s\n", completed, milestone)
	return nil
}

// routeAction sends one action to the dispatcher owning its service —
// shared by every chain command (pr, gate).
func routeAction(c *Context, a action.Action) error {
	switch a.Service {
	case "sheet":
		return sheetDispatch(c, a)
	case "wecom":
		return wecomDispatch(c, a)
	default:
		return fmt.Errorf("no dispatcher for service %q", a.Service)
	}
}
