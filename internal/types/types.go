// Package types defines the data structures shared by rules, the corrector and
// the user interface.
package types

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"
	"sync"

	"github.com/leafevans/go-fuck/internal/conf"
	"github.com/leafevans/go-fuck/internal/logs"
	"github.com/leafevans/go-fuck/internal/output"
	"github.com/leafevans/go-fuck/internal/shells"
)

// ErrEmptyCommand is returned when an empty command is passed to go-fuck.
var ErrEmptyCommand = errors.New("empty command")

// Command is a command that should be fixed.
type Command struct {
	// Script is the command line, with shell aliases already expanded.
	Script string
	// Output is the combined standard output and standard error of the
	// command. It is only meaningful when HasOutput reports true.
	Output string

	hasOutput bool

	partsOnce sync.Once
	parts     []string
}

// NewCommand returns a command with a known output.
func NewCommand(script, output string) *Command {
	return &Command{Script: script, Output: output, hasOutput: true}
}

// NewCommandWithoutOutput returns a command whose output could not be
// obtained, for example because re-running it timed out.
func NewCommandWithoutOutput(script string) *Command {
	return &Command{Script: script}
}

// FromRawScript creates a command from the command line arguments that were
// passed to go-fuck.
func FromRawScript(raw []string) (*Command, error) {
	script := FormatRawScript(raw)
	if script == "" {
		return nil, ErrEmptyCommand
	}

	expanded := shells.Current().FromShell(script)
	commandOutput, ok := output.GetOutput(script, expanded)
	return &Command{Script: expanded, Output: commandOutput, hasOutput: ok}, nil
}

// FormatRawScript joins the raw command line parts into a single command.
func FormatRawScript(raw []string) string {
	return strings.TrimLeft(strings.Join(raw, " "), " ")
}

// HasOutput reports whether the command output was captured.
func (c *Command) HasOutput() bool {
	return c.hasOutput
}

// ScriptParts returns the command split into arguments using shell rules.
// Unparsable commands are split on whitespace.
func (c *Command) ScriptParts() []string {
	c.partsOnce.Do(func() {
		c.parts = shells.Current().SplitCommand(c.Script)
	})
	return c.parts
}

// Update returns a copy of the command with the script and output replaced.
func (c *Command) Update(script, commandOutput string) *Command {
	return &Command{Script: script, Output: commandOutput, hasOutput: c.hasOutput}
}

// WithScript returns a copy of the command with the script replaced.
func (c *Command) WithScript(script string) *Command {
	return c.Update(script, c.Output)
}

// WithOutput returns a copy of the command with the output replaced.
func (c *Command) WithOutput(commandOutput string) *Command {
	return c.Update(c.Script, commandOutput)
}

// String implements fmt.Stringer.
func (c *Command) String() string {
	return fmt.Sprintf("Command(script=%s, output=%s)", c.Script, c.Output)
}

// MatchFunc reports whether a rule matches a command.
type MatchFunc func(*Command) bool

// NewCommandFunc returns the corrected commands for a matched command.
type NewCommandFunc func(*Command) ([]string, error)

// SideEffectFunc performs a side effect before a corrected command runs.
type SideEffectFunc func(oldCommand *Command, script string) error

// Rule fixes commands.
type Rule struct {
	// Name is the rule name, derived from its file name.
	Name string
	// Match reports whether the rule applies to a command.
	Match MatchFunc
	// GetNewCommand returns one or more corrected commands.
	GetNewCommand NewCommandFunc
	// EnabledByDefault reports whether the rule is enabled unless it is listed
	// in the rules setting.
	EnabledByDefault bool
	// SideEffect is applied before the corrected command runs.
	SideEffect SideEffectFunc
	// Priority orders the rules. Lower values are tried first.
	Priority int
	// RequiresOutput reports whether the rule needs the command output.
	RequiresOutput bool
}

// NewRule returns a rule with the default flags.
func NewRule(name string, match MatchFunc, getNewCommand NewCommandFunc) *Rule {
	return &Rule{
		Name:             name,
		Match:            match,
		GetNewCommand:    getNewCommand,
		EnabledByDefault: true,
		Priority:         1000,
		RequiresOutput:   true,
	}
}

// IsEnabled reports whether the rule is enabled by the settings.
func (r *Rule) IsEnabled() bool {
	for _, name := range conf.Current().Rules {
		if name == r.Name {
			return true
		}
	}
	if !r.EnabledByDefault {
		return false
	}
	for _, name := range conf.Current().Rules {
		if name == "All rules enabled" {
			return true
		}
	}
	return false
}

// IsMatch reports whether the rule matches the command. A panicking or
// erroring rule is reported and treated as a non-match.
func (r *Rule) IsMatch(command *Command) (matched bool) {
	if !command.HasOutput() && r.RequiresOutput {
		return false
	}

	stopTimer := logs.DebugTime("Trying rule: " + r.Name + ";")
	defer stopTimer()
	defer func() {
		if recovered := recover(); recovered != nil {
			logs.RuleFailed(r.Name, fmt.Errorf("%v", recovered))
			matched = false
		}
	}()

	return r.Match(command)
}

// GetCorrectedCommands returns corrected commands for a matched command.
//
// Priorities of multiple corrections are spread so that the first suggestion
// has the rule priority, the second one twice the rule priority and so on.
func (r *Rule) GetCorrectedCommands(command *Command) (corrected []*CorrectedCommand) {
	defer func() {
		if recovered := recover(); recovered != nil {
			logs.RuleFailed(r.Name, fmt.Errorf("%v", recovered))
			corrected = nil
		}
	}()

	newCommands, err := r.GetNewCommand(command)
	if err != nil {
		logs.RuleFailed(r.Name, err)
		return nil
	}

	for index, script := range newCommands {
		corrected = append(corrected, &CorrectedCommand{
			Script:     script,
			SideEffect: r.SideEffect,
			Priority:   (index + 1) * r.Priority,
		})
	}
	return corrected
}

// String implements fmt.Stringer.
func (r *Rule) String() string {
	return fmt.Sprintf("Rule(name=%s, priority=%d, enabled_by_default=%t, requires_output=%t)",
		r.Name, r.Priority, r.EnabledByDefault, r.RequiresOutput)
}

// CorrectedCommand is a command fixed by a rule.
type CorrectedCommand struct {
	// Script is the corrected command line.
	Script string
	// SideEffect is executed before the corrected command runs.
	SideEffect SideEffectFunc
	// Priority orders the corrected commands. Lower values are suggested
	// first.
	Priority int
}

// Equal reports whether two corrected commands are equal. The priority is
// ignored, mirroring the original implementation.
func (c *CorrectedCommand) Equal(other *CorrectedCommand) bool {
	if c == nil || other == nil {
		return c == other
	}
	return c.Script == other.Script && sameSideEffect(c.SideEffect, other.SideEffect)
}

// sameSideEffect reports whether two side effect functions are the same.
func sameSideEffect(a, b SideEffectFunc) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return reflect.ValueOf(a).Pointer() == reflect.ValueOf(b).Pointer()
}

// String implements fmt.Stringer.
func (c *CorrectedCommand) String() string {
	return fmt.Sprintf("CorrectedCommand(script=%s, side_effect=%t, priority=%d)",
		c.Script, c.SideEffect != nil, c.Priority)
}

// runnableScript returns the script that is executed in the shell.
//
// When the repeat setting is enabled, the script is followed by a second
// go-fuck call that fixes the command again if it fails.
func (c *CorrectedCommand) runnableScript() string {
	if !conf.Current().Repeat {
		return c.Script
	}
	debug := ""
	if conf.Current().Debug {
		debug = "--debug "
	}
	repeat := fmt.Sprintf("%s --repeat %s--force-command %s",
		conf.Alias(), debug, shells.Current().Quote(c.Script))
	return shells.Current().Or(c.Script, repeat)
}

// Run applies the side effect, updates the shell history and prints the
// corrected command to standard output.
func (c *CorrectedCommand) Run(oldCommand *Command) {
	if c.SideEffect != nil {
		if err := c.SideEffect(oldCommand, c.Script); err != nil {
			logs.RuleFailed(c.Script, err)
		}
	}
	if conf.Current().AlterHistory {
		if err := shells.Current().PutToHistory(c.Script); err != nil {
			logs.Exception("Can't update history", err)
		}
	}
	fmt.Fprint(os.Stdout, c.runnableScript())
}
