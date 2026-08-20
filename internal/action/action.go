// Package action defines the external side-effect unit that flows
// through fbh. Every call to Tencent sheet, WeCom, or GitHub is
// described as an Action; with --dry-run the Action is printed as one
// JSON line instead of being dispatched.
package action

// Action describes one external side effect fbh is about to perform.
type Action struct {
	Service string         `json:"service"` // "sheet" | "gh"
	Op      string         `json:"op"`
	Payload map[string]any `json:"payload,omitempty"`
}
