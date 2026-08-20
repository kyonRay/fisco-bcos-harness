// Package mcp is a minimal MCP streamable-HTTP client, just enough to
// call Tencent Docs tools. The auth token is NOT managed here: fbh
// reuses the token the tencent-docs skill's auth flow stored in
// mcporter's config (ADR-0006).
package mcp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// DefaultBaseURL is the production Tencent Docs MCP endpoint.
const DefaultBaseURL = "https://docs.qq.com/openapi/mcp"

// Client calls MCP tools over streamable HTTP.
type Client struct {
	BaseURL string
	Token   string
	HTTP    *http.Client

	sessionID   string
	initialized bool
}

// LoadToken reads the Tencent Docs Authorization token that the
// tencent-docs skill's auth flow saved into mcporter's config.
// Path override: $FBH_MCPORTER_CONFIG (tests), else ~/.mcporter/mcporter.json.
func LoadToken() (string, error) {
	path := os.Getenv("FBH_MCPORTER_CONFIG")
	if path == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		path = filepath.Join(home, ".mcporter", "mcporter.json")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read mcporter config (run the tencent-docs auth flow via /fbh-setup first): %w", err)
	}
	var cfg struct {
		MCPServers map[string]struct {
			Headers map[string]string `json:"headers"`
		} `json:"mcpServers"`
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return "", fmt.Errorf("parse %s: %w", path, err)
	}
	srv, ok := cfg.MCPServers["tencent-docs"]
	if !ok || srv.Headers["Authorization"] == "" {
		return "", fmt.Errorf("no tencent-docs token in %s; run the tencent-docs auth flow via /fbh-setup", path)
	}
	return srv.Headers["Authorization"], nil
}

type rpcRequest struct {
	JSONRPC string `json:"jsonrpc"`
	ID      *int   `json:"id,omitempty"`
	Method  string `json:"method"`
	Params  any    `json:"params,omitempty"`
}

type rpcResponse struct {
	Result struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
		IsError bool `json:"isError"`
	} `json:"result"`
	Error *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func (c *Client) post(req rpcRequest) (*http.Response, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	httpReq, err := http.NewRequest(http.MethodPost, c.BaseURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json, text/event-stream")
	httpReq.Header.Set("Authorization", c.Token)
	if c.sessionID != "" {
		httpReq.Header.Set("Mcp-Session-Id", c.sessionID)
	}
	client := c.HTTP
	if client == nil {
		client = &http.Client{Timeout: 60 * time.Second}
		c.HTTP = client
	}
	return client.Do(httpReq)
}

func (c *Client) handshake() error {
	one := 1
	resp, err := c.post(rpcRequest{
		JSONRPC: "2.0", ID: &one, Method: "initialize",
		Params: map[string]any{
			"protocolVersion": "2025-03-26",
			"clientInfo":      map[string]any{"name": "fbh", "version": "0.1"},
			"capabilities":    map[string]any{},
		},
	})
	if err != nil {
		return fmt.Errorf("mcp initialize: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("mcp initialize: HTTP %d", resp.StatusCode)
	}
	if sid := resp.Header.Get("Mcp-Session-Id"); sid != "" {
		c.sessionID = sid
	}
	notif, err := c.post(rpcRequest{JSONRPC: "2.0", Method: "notifications/initialized"})
	if err != nil {
		return fmt.Errorf("mcp initialized notification: %w", err)
	}
	notif.Body.Close()
	c.initialized = true
	return nil
}

// CallTool performs the MCP handshake (once) and invokes one tool,
// returning the concatenated text content of the result.
func (c *Client) CallTool(name string, args map[string]any) (string, error) {
	if !c.initialized {
		if err := c.handshake(); err != nil {
			return "", err
		}
	}
	two := 2
	resp, err := c.post(rpcRequest{
		JSONRPC: "2.0", ID: &two, Method: "tools/call",
		Params: map[string]any{"name": name, "arguments": args},
	})
	if err != nil {
		return "", fmt.Errorf("mcp tools/call %s: %w", name, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("mcp tools/call %s: HTTP %d", name, resp.StatusCode)
	}
	payload, err := decodeBody(resp)
	if err != nil {
		return "", fmt.Errorf("mcp tools/call %s: %w", name, err)
	}
	var rpc rpcResponse
	if err := json.Unmarshal(payload, &rpc); err != nil {
		return "", fmt.Errorf("mcp tools/call %s: parse response: %w", name, err)
	}
	if rpc.Error != nil {
		return "", fmt.Errorf("mcp tools/call %s: %s (code %d)", name, rpc.Error.Message, rpc.Error.Code)
	}
	var texts []string
	for _, item := range rpc.Result.Content {
		if item.Type == "text" {
			texts = append(texts, item.Text)
		}
	}
	joined := strings.Join(texts, "\n")
	if rpc.Result.IsError {
		return "", fmt.Errorf("mcp tool %s returned error: %s", name, joined)
	}
	return joined, nil
}

// decodeBody handles both plain JSON and SSE ("data: {...}") response
// bodies, returning the JSON payload bytes.
func decodeBody(resp *http.Response) ([]byte, error) {
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(resp.Body); err != nil {
		return nil, err
	}
	raw := buf.Bytes()
	if strings.HasPrefix(resp.Header.Get("Content-Type"), "text/event-stream") {
		for _, line := range strings.Split(string(raw), "\n") {
			if strings.HasPrefix(line, "data:") {
				return []byte(strings.TrimSpace(strings.TrimPrefix(line, "data:"))), nil
			}
		}
		return nil, fmt.Errorf("no data line in SSE response")
	}
	return raw, nil
}
