package cli

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/kyonRay/fisco-bcos-harness/internal/action"
)

// fakeCmd registers a subcommand that emits one wecom action and counts
// real dispatches, returning the counter.
func registerFakeCmd(t *testing.T, name string) *int {
	t.Helper()
	dispatched := 0
	Register(Command{
		Name:    name,
		Summary: "test-only fake command",
		Exec: func(c *Context) error {
			return c.Do(action.Action{
				Service: "wecom",
				Op:      "nudge",
				Payload: map[string]any{"pr": "https://example.com/pr/42"},
			})
		},
		Dispatch: func(a action.Action) error {
			dispatched++
			return nil
		},
	})
	return &dispatched
}

func TestDryRunPrintsActionJSONWithoutDispatching(t *testing.T) {
	dispatched := registerFakeCmd(t, "fake-dry")
	var stdout, stderr bytes.Buffer

	code := Run([]string{"fake-dry", "--dry-run"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", code, stderr.String())
	}
	if *dispatched != 0 {
		t.Fatalf("dispatched %d actions in dry-run, want 0", *dispatched)
	}
	line := strings.TrimSpace(stdout.String())
	var got action.Action
	if err := json.Unmarshal([]byte(line), &got); err != nil {
		t.Fatalf("stdout is not one action JSON line: %q: %v", line, err)
	}
	if got.Service != "wecom" || got.Op != "nudge" {
		t.Fatalf("action = %+v, want service=wecom op=nudge", got)
	}
	if got.Payload["pr"] != "https://example.com/pr/42" {
		t.Fatalf("payload = %v, want pr link preserved", got.Payload)
	}
}

func TestWithoutDryRunDispatchesAndPrintsNoActionJSON(t *testing.T) {
	dispatched := registerFakeCmd(t, "fake-real")
	var stdout, stderr bytes.Buffer

	code := Run([]string{"fake-real"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", code, stderr.String())
	}
	if *dispatched != 1 {
		t.Fatalf("dispatched = %d, want 1", *dispatched)
	}
	if strings.Contains(stdout.String(), "\"service\"") {
		t.Fatalf("real run must not print action JSON, got: %s", stdout.String())
	}
}
