// Package ui implements the interactive selection of a corrected command.
package ui

import (
	"github.com/leafevans/go-fuck/v4/internal/conf"
	"github.com/leafevans/go-fuck/v4/internal/consts"
	"github.com/leafevans/go-fuck/v4/internal/logs"
	"github.com/leafevans/go-fuck/v4/internal/system"
	"github.com/leafevans/go-fuck/v4/internal/types"
)

// newline is the line separator.
const newline = "\n"

// ReadKey returns the next key press. It is replaceable in tests.
var ReadKey = system.GetKey

// ReadActions returns a function that yields the user action for each pressed
// key. It returns an abort action when a key cannot be read.
func ReadActions() func() consts.Action {
	return func() consts.Action {
		for {
			key, err := ReadKey()
			if err != nil {
				return consts.ActionAbort
			}
			switch key {
			case consts.KeyUp, consts.KeyCtrlN, "k", "e":
				return consts.ActionPrevious
			case consts.KeyDown, consts.KeyCtrlP, "j", "n":
				return consts.ActionNext
			case consts.KeyCtrlC, "q":
				return consts.ActionAbort
			case "\n", "\r":
				return consts.ActionSelect
			}
		}
	}
}

// CommandSelector selects one of the corrected commands.
type CommandSelector struct {
	commands []*types.CorrectedCommand
	index    int
}

// NewCommandSelector returns a selector for the corrected commands. The second
// result is false when there is nothing to select.
func NewCommandSelector(commands []*types.CorrectedCommand) (*CommandSelector, bool) {
	if len(commands) == 0 {
		return nil, false
	}
	return &CommandSelector{commands: commands}, true
}

// Next selects the next command, wrapping around.
func (s *CommandSelector) Next() {
	s.index = (s.index + 1) % len(s.commands)
}

// Previous selects the previous command, wrapping around.
func (s *CommandSelector) Previous() {
	s.index = (s.index - 1 + len(s.commands)) % len(s.commands)
}

// Value returns the selected command.
func (s *CommandSelector) Value() *types.CorrectedCommand {
	return s.commands[s.index]
}

// confirmSelection prints the confirmation prompt for a command.
func confirmSelection(selector *CommandSelector) {
	value := selector.Value()
	logs.ConfirmText(value.Script, value.SideEffect != nil, value.Explanation(), value.BeforeRun())
}

// PreviewCommands renders suggestions without selecting one or writing shell code.
func PreviewCommands(commands []*types.CorrectedCommand) {
	logs.PreviewHeader()
	for _, command := range commands {
		showCommand(command)
	}
}

// showCommand renders the currently highlighted suggestion to stderr.
func showCommand(command *types.CorrectedCommand) {
	logs.ShowCorrectedCommand(command.Script, command.SideEffect != nil, command.Explanation(), command.BeforeRun())
}

// SelectCommand returns the command chosen by the user.
//
// The first suggestion is returned when confirmation is disabled. A nil result
// means that no rule matched or that the user aborted the selection.
func SelectCommand(commands []*types.CorrectedCommand) *types.CorrectedCommand {
	selector, ok := NewCommandSelector(commands)
	if !ok {
		return nil
	}

	if !conf.Current().RequireConfirmation {
		value := selector.Value()
		showCommand(value)
		return value
	}

	if conf.Current().Fzf {
		if selected, handled := selectWithFzf(commands); handled {
			return selected
		}
	}

	// The Windows key reader switches the console out of line input mode; make
	// sure the parent shell gets its normal input handling back.
	defer system.RestoreInput()
	confirmSelection(selector)
	nextAction := ReadActions()
	for {
		switch nextAction() {
		case consts.ActionSelect:
			logs.NewLine()
			return selector.Value()
		case consts.ActionAbort:
			logs.NewLine()
			logs.Failed("Aborted")
			return nil
		case consts.ActionPrevious:
			selector.Previous()
			confirmSelection(selector)
		case consts.ActionNext:
			selector.Next()
			confirmSelection(selector)
		}
	}
}
