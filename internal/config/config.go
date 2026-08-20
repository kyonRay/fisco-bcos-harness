// Package config loads and stores the per-developer fbh configuration.
//
// Credentials policy (ADR-0006): the Tencent Docs token is NOT stored
// here — fbh reuses the token the tencent-docs skill's auth flow saved
// into mcporter's config. fbh's own file only holds the team sheet id
// and the shared WeCom webhook key, and never enters any git repo.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config is the on-disk fbh configuration.
type Config struct {
	// SheetFileID is the file_id of the team's Tencent smart sheet
	// (the requirement-status source of truth).
	SheetFileID string `json:"sheet_file_id,omitempty"`
	// WecomWebhook is the team-shared WeCom group bot webhook URL.
	WecomWebhook string `json:"wecom_webhook,omitempty"`
	// MCPBaseURL overrides the Tencent Docs MCP endpoint (tests use
	// this to point at a fake server). Empty means the production URL.
	MCPBaseURL string `json:"mcp_base_url,omitempty"`
	// GateCmd is the shell command that runs the milestone gate (the
	// fisco-bcos-release-gate invocation); fbh wraps it, never
	// reimplements it (ADR-0004).
	GateCmd string `json:"gate_cmd,omitempty"`
}

// Path returns the config file location: $FBH_CONFIG if set, else
// ~/.fbh/config.json.
func Path() string {
	if p := os.Getenv("FBH_CONFIG"); p != "" {
		return p
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ".fbh-config.json"
	}
	return filepath.Join(home, ".fbh", "config.json")
}

// Load reads the config file at Path. A missing file returns an empty
// Config and no error — callers decide which fields are required.
func Load() (Config, error) {
	var cfg Config
	data, err := os.ReadFile(Path())
	if os.IsNotExist(err) {
		return cfg, nil
	}
	if err != nil {
		return cfg, err
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("parse %s: %w", Path(), err)
	}
	return cfg, nil
}

// Save writes the config file at Path with owner-only permissions,
// creating the parent directory if needed.
func Save(cfg Config) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	path := Path()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o600)
}
