// Package ui implements the interactive selection of a corrected command.
package ui

import (
	"fmt"
	"os"

	"github.com/leafevans/go-fuck/internal/conf"
	"github.com/leafevans/go-fuck/internal/consts"
	"github.com/leafevans/go-fuck/internal/logs"
	"github.com/leafevans/go-fuck/internal/system"
	"github.com/leafevans/go-fuck/internal/types"
)

// newline is the line separator.
var newline = string(rune(10))

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
			switch {
			case key == consts.KeyUp || key == consts.KeyCtrlN || key == "k" || key == "e":
				return consts.ActionPrevious
			case key == consts.KeyDown || key == consts.KeyCtrlP || key == "j" || key == "n":
				return consts.ActionNext
			case key == consts.KeyCtrlC || key == "q":
				return consts.ActionAbort
			case string(key) == newline || string(key) == string(rune(13)):
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
	logs.ConfirmText(value.Script, value.SideEffect != nil)
}

// SelectCommand returns the command chosen by the user.
//
// The first suggestion is returned when confirmation is disabled. A nil result
// means that no rule matched or that the user aborted the selection.
func SelectCommand(commands []*types.CorrectedCommand) *types.CorrectedCommand {
	selector, ok := NewCommandSelector(commands)
	if !ok {
		if conf.Alias() == "fuck" {
			logs.Failed("No fucks given")
		} else {
			logs.Failed("Nothing found")
		}
		return nil
	}

	if !conf.Current().RequireConfirmation {
		value := selector.Value()
		logs.ShowCorrectedCommand(value.Script, value.SideEffect != nil)
		return value
	}

	confirmSelection(selector)
	nextAction := ReadActions()
	for {
		switch nextAction() {
		case consts.ActionSelect:
			fmt.Fprint(os.Stderr, newline)
			return selector.Value()
		case consts.ActionAbort:
			logs.Failed(newline + "Aborted")
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
