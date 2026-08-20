package cli

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// pagedMCPServer serves list_records in two pages so pagination is
// observable: the target record lives on page 2.
func pagedMCPServer(t *testing.T) (*httptest.Server, *[]map[string]any) {
	t.Helper()
	var updated []map[string]any
	page1 := `{"total":2,"has_more":true,"next":1,"records":[
	  {"record_id":"p1","field_values":[{"field":"需求名","text_value":{"items":[{"type":"text","text":"别的需求"}]}}]}]}`
	page2 := `{"total":2,"has_more":false,"records":[
	  {"record_id":"p2","field_values":[{"field":"需求名","text_value":{"items":[{"type":"text","text":"登录需求"}]}}]}]}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			ID     any    `json:"id"`
			Method string `json:"method"`
			Params struct {
				Name      string         `json:"name"`
				Arguments map[string]any `json:"arguments"`
			} `json:"params"`
		}
		json.NewDecoder(r.Body).Decode(&req)
		w.Header().Set("Content-Type", "application/json")
		reply := func(text string) {
			json.NewEncoder(w).Encode(map[string]any{
				"jsonrpc": "2.0", "id": req.ID,
				"result": map[string]any{"content": []map[string]any{{"type": "text", "text": text}}},
			})
		}
		switch req.Method {
		case "initialize":
			json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": req.ID, "result": map[string]any{}})
		case "notifications/initialized":
			w.WriteHeader(http.StatusAccepted)
		case "tools/call":
			switch req.Params.Name {
			case "smartsheet.list_records":
				if off, _ := req.Params.Arguments["offset"].(float64); off > 0 {
					reply(page2)
				} else {
					reply(page1)
				}
			case "smartsheet.update_records":
				updated = append(updated, req.Params.Arguments)
				reply(`{"records":[{"record_id":"p2"}]}`)
			case "smartsheet.add_records":
				t.Errorf("must update the page-2 record, not add a duplicate")
				reply(`{}`)
			}
		}
	}))
	return srv, &updated
}

func TestUpsertFindsRecordBeyondFirstPage(t *testing.T) {
	srv, updated := pagedMCPServer(t)
	defer srv.Close()
	setupSheetEnv(t, srv.URL)
	var stdout, stderr bytes.Buffer

	code := Run([]string{"sheet", "set-status", "--table", "s1", "--key", "登录需求", "--status", "开发中"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit = %d, stderr: %s", code, stderr.String())
	}
	if len(*updated) != 1 {
		t.Fatalf("updated = %d, want 1 (record found on page 2)", len(*updated))
	}
}

func TestClaimByCurrentOwnerIsNoOp(t *testing.T) {
	srv := &recordsMCPServer{existing: loginRecordOwned} // 认领人=李四
	ts := srv.start(t)
	defer ts.Close()
	setupSheetEnv(t, ts.URL)
	var stdout, stderr bytes.Buffer

	code := Run([]string{"sheet", "claim", "--table", "s1", "--key", "登录需求", "--owner", "李四"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("same-owner re-claim must succeed, exit = %d: %s", code, stderr.String())
	}
	if len(srv.updated)+len(srv.added) != 0 {
		t.Fatalf("same-owner re-claim must not write (would drag status back), writes = %d",
			len(srv.updated)+len(srv.added))
	}
}

func TestGhMyReviewsSubcommandListsLoopStates(t *testing.T) {
	ghEnv(t)
	requested, reviewed := reviewFixtures()
	standupStub(t, "[]", requested, reviewed)
	var stdout, stderr bytes.Buffer

	code := Run([]string{"gh", "my-reviews"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit = %d, stderr: %s", code, stderr.String())
	}
	out := stdout.String()
	for _, want := range []string{"pull/51", "等我首审", "pull/54", "等人工approve"} {
		if !strings.Contains(out, want) {
			t.Fatalf("my-reviews missing %q, got:\n%s", want, out)
		}
	}
}
