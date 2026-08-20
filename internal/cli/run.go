// Package cli implements the fbh command-line entry point.
//
// fbh is the single seam through which harness skills touch external
// services (Tencent smart sheet, WeCom webhook, GitHub). Skills never
// call those services directly.
package cli

import (
	"fmt"
	"io"
	"strings"
)

// Version is the fbh version string. Overridable at build time via
// -ldflags "-X .../internal/cli.Version=vX.Y.Z".
var Version = "v0.1.0"

// Run executes fbh with the given arguments and returns the process
// exit code. args excludes the program name. --dry-run may appear
// anywhere in the arguments.
func Run(args []string, stdout, stderr io.Writer) int {
	dryRun := false
	rest := make([]string, 0, len(args))
	for _, arg := range args {
		if arg == "--dry-run" {
			dryRun = true
			continue
		}
		rest = append(rest, arg)
	}

	if len(rest) == 0 {
		fmt.Fprintf(stderr, "usage: fbh [--dry-run] <command>\n")
		return 2
	}
	if rest[0] == "--version" {
		fmt.Fprintf(stdout, "fbh %s\n", Version)
		return 0
	}

	cmd, ok := registry[rest[0]]
	if !ok {
		fmt.Fprintf(stderr, "fbh: unknown command %q\navailable commands: %s\n",
			rest[0], strings.Join(commandNames(), ", "))
		return 2
	}
	ctx := &Context{
		Stdout:   stdout,
		Stderr:   stderr,
		Args:     rest[1:],
		dryRun:   dryRun,
		dispatch: cmd.Dispatch,
	}
	if err := cmd.Exec(ctx); err != nil {
		fmt.Fprintf(stderr, "fbh %s: %v\n", cmd.Name, err)
		return 1
	}
	return 0
}
