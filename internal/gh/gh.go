// Package gh shells out to the GitHub CLI. Tests point FBH_GH_BIN at a
// stub script to replay fixtures; production uses the real `gh`.
package gh

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func bin() string {
	if b := os.Getenv("FBH_GH_BIN"); b != "" {
		return b
	}
	return "gh"
}

// Run executes gh with args and returns trimmed stdout.
func Run(args ...string) (string, error) {
	cmd := exec.Command(bin(), args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("gh %s: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(stderr.String()))
	}
	return strings.TrimSpace(stdout.String()), nil
}
