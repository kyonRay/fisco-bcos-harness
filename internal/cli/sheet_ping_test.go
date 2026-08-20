package cli

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// fakeMCPServer implements just enough of the MCP streamable-HTTP
// protocol for sheet ping: initialize, notifications/initialized, and
// tools/call for smartsheet.list_tables.
func fakeMCPServer(t *testing.T, wantToken string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != wantToken {
			t.Errorf("Authorization = %q, want %q", got, wantToken)
			http.Error(w, "bad token", http.StatusUnauthorized)
			return
		}
		var req struct {
			ID     any    `json:"id"`
			Method string `json:"method"`
			Params struct {
				Name      string         `json:"name"`
				Arguments map[string]any `json:"arguments"`
			} `json:"params"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("bad request body: %v", err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		switch req.Method {
		case "initialize":
			w.Header().Set("Mcp-Session-Id", "sess-1")
			resp := map[string]any{"jsonrpc": "2.0", "id": req.ID, "result": map[string]any{}}
			json.NewEncoder(w).Encode(resp)
		case "notifications/initialized":
			w.WriteHeader(http.StatusAccepted)
		case "tools/call":
			if r.Header.Get("Mcp-Session-Id") != "sess-1" {
				t.Errorf("tools/call missing session id")
			}
			if req.Params.Name != "smartsheet.list_tables" {
				t.Errorf("tool = %q, want smartsheet.list_tables", req.Params.Name)
			}
			if req.Params.Arguments["file_id"] != "SHEET123" {
				t.Errorf("file_id = %v, want SHEET123", req.Params.Arguments["file_id"])
			}
			text := `{"sheets":[{"sheet_id":"s1","name":"需求表"},{"sheet_id":"s2","name":"milestone表"}]}`
			resp := map[string]any{
				"jsonrpc": "2.0", "id": req.ID,
				"result": map[string]any{
					"content": []map[string]any{{"type": "text", "text": text}},
				},
			}
			json.NewEncoder(w).Encode(resp)
		default:
			t.Errorf("unexpected method %q", req.Method)
		}
	}))
}

func writeFakeMcporter(t *testing.T, token string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "mcporter.json")
	body := `{"mcpServers":{"tencent-docs":{"baseUrl":"https://docs.qq.com/openapi/mcp","headers":{"Authorization":"` + token + `"}}}}`
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestSheetPingConfiguredListsTableNames(t *testing.T) {
	srv := fakeMCPServer(t, "Bearer FAKETOKEN")
	defer srv.Close()

	cfgPath := filepath.Join(t.TempDir(), "config.json")
	t.Setenv("FBH_CONFIG", cfgPath)
	t.Setenv("FBH_MCPORTER_CONFIG", writeFakeMcporter(t, "Bearer FAKETOKEN"))
	cfg := `{"sheet_file_id":"SHEET123","mcp_base_url":"` + srv.URL + `"}`
	if err := os.WriteFile(cfgPath, []byte(cfg), 0o600); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	code := Run([]string{"sheet", "ping"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit = %d, stderr: %s", code, stderr.String())
	}
	out := stdout.String()
	if !strings.Contains(out, "需求表") || !strings.Contains(out, "milestone表") {
		t.Fatalf("ping must echo table names, got: %s", out)
	}
}
