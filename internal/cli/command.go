package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"

	"github.com/kyonRay/fisco-bcos-harness/internal/action"
)

// Command is one fbh subcommand. Exec describes what to do by emitting
// Actions through Context.Do; Dispatch performs one Action for real.
// The dry-run split lives here so every future subcommand gets it for free.
type Command struct {
	Name     string
	Summary  string
	Exec     func(*Context) error
	Dispatch func(action.Action) error
}

// Context is passed to a Command's Exec.
type Context struct {
	Stdout io.Writer
	Stderr io.Writer
	Args   []string // arguments after the subcommand name, flags stripped

	dryRun   bool
	dispatch func(action.Action) error
}

// Do routes one Action: printed as a JSON line in dry-run mode,
// dispatched for real otherwise.
func (c *Context) Do(a action.Action) error {
	if c.dryRun {
		line, err := json.Marshal(a)
		if err != nil {
			return fmt.Errorf("encode action: %w", err)
		}
		fmt.Fprintln(c.Stdout, string(line))
		return nil
	}
	return c.dispatch(a)
}

var registry = map[string]Command{}

// Register adds a subcommand to fbh. Later tickets register the real
// sheet/wecom/gh commands; tests register fakes.
func Register(cmd Command) {
	registry[cmd.Name] = cmd
}

func commandNames() []string {
	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
