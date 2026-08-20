package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/kyonRay/fisco-bcos-harness/internal/action"
	"github.com/kyonRay/fisco-bcos-harness/internal/config"
)

func init() {
	Register(Command{
		Name:     "wecom",
		Summary:  "send directed WeCom group-bot mentions (nudges)",
		Exec:     wecomExec,
		Dispatch: wecomDispatch,
	})
}

func wecomExec(c *Context) error {
	if len(c.Args) == 0 || c.Args[0] != "nudge" {
		return fmt.Errorf("usage: fbh wecom nudge --to <成员> --pr <PR链接> [--text <说明>]")
	}
	flags, _, err := parseFlags(c.Args[1:])
	if err != nil {
		return err
	}
	if flags["to"] == "" || flags["pr"] == "" {
		return fmt.Errorf("usage: fbh wecom nudge --to <成员> --pr <PR链接> [--text <说明>]")
	}
	return c.Do(nudgeAction(flags["to"], flags["pr"], flags["text"]))
}

func nudgeAction(to, pr, text string) action.Action {
	if text == "" {
		text = "请 review"
	}
	return action.Action{
		Service: "wecom",
		Op:      "nudge",
		Payload: map[string]any{"to": to, "pr": pr, "text": text},
	}
}

func wecomDispatch(c *Context, a action.Action) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	if cfg.WecomWebhook == "" {
		return fmt.Errorf("wecom_webhook is not configured; run the /fbh-setup skill first")
	}
	to := payloadStr(a.Payload, "to")
	msg := map[string]any{
		"msgtype": "text",
		"text": map[string]any{
			"content":        fmt.Sprintf("%s：%s %s", to, payloadStr(a.Payload, "text"), payloadStr(a.Payload, "pr")),
			"mentioned_list": []string{to},
		},
	}
	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	resp, err := http.Post(cfg.WecomWebhook, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("wecom webhook: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("wecom webhook: HTTP %d", resp.StatusCode)
	}
	fmt.Fprintf(c.Stdout, "nudged %s about %s\n", to, payloadStr(a.Payload, "pr"))
	return nil
}
