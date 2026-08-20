package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

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
	// "to" may be a comma-separated list; each login goes through the
	// mention_map so the WeCom @ actually lands.
	var mentionMap map[string]string
	if cfg.MentionMap != "" {
		if err := json.Unmarshal([]byte(cfg.MentionMap), &mentionMap); err != nil {
			return fmt.Errorf("parse mention_map config: %w", err)
		}
	}
	var mentions []string
	for _, login := range strings.Split(payloadStr(a.Payload, "to"), ",") {
		login = strings.TrimSpace(login)
		if login == "" {
			continue
		}
		if mapped := mentionMap[login]; mapped != "" {
			login = mapped
		}
		mentions = append(mentions, login)
	}
	msg := map[string]any{
		"msgtype": "text",
		"text": map[string]any{
			"content":        fmt.Sprintf("%s：%s %s", strings.Join(mentions, " "), payloadStr(a.Payload, "text"), payloadStr(a.Payload, "pr")),
			"mentioned_list": mentions,
		},
	}
	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Post(cfg.WecomWebhook, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("wecom webhook: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("wecom webhook: HTTP %d", resp.StatusCode)
	}
	fmt.Fprintf(c.Stdout, "nudged %s about %s\n", strings.Join(mentions, ","), payloadStr(a.Payload, "pr"))
	return nil
}
