// Package ui implements the interactive selection of a corrected command.
package ui

import (
	"errors"
	"os"
	"os/exec"
	"strings"

	"github.com/leafevans/go-fuck/internal/conf"
	"github.com/leafevans/go-fuck/internal/consts"
	"github.com/leafevans/go-fuck/internal/logs"
	"github.com/leafevans/go-fuck/internal/system"
	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
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

// errFzfUnavailable is returned when the fzf executable is not installed.
var errFzfUnavailable = errors.New("fzf is not available")

// errFzfAborted is returned when the user dismisses the fzf prompt.
var errFzfAborted = errors.New("fzf selection aborted")

// FzfRunner runs fzf with the candidate commands and returns the selected one.
// It is a variable so that tests can stub the fuzzy finder.
var FzfRunner = runFzf

// runFzf starts fzf with the candidate commands on standard input.
func runFzf(candidates []string) (string, error) {
	fzf := utils.Which("fzf")
	if fzf == "" {
		return "", errFzfUnavailable
	}

	command := exec.Command(fzf,
		"--height=40%",
		"--reverse",
		"--no-multi",
		"--prompt=fuck> ",
		"--header=enter: select   esc: abort")
	command.Stdin = strings.NewReader(strings.Join(candidates, newline) + newline)
	command.Stderr = os.Stderr
	output, err := command.Output()
	if err != nil {
		var exitError *exec.ExitError
		if errors.As(err, &exitError) && (exitError.ExitCode() == 1 || exitError.ExitCode() == 130) {
			return "", errFzfAborted
		}
		return "", err
	}
	return strings.TrimRight(string(output), newline+"\r"), nil
}

// selectWithFzf asks fzf to select one of the corrected commands.
//
// The second result reports whether the selection was handled by fzf. When it
// is false the caller should fall back to the built-in prompt.
func selectWithFzf(commands []*types.CorrectedCommand) (*types.CorrectedCommand, bool) {
	candidates := make([]string, 0, len(commands))
	byScript := make(map[string]*types.CorrectedCommand, len(commands))
	for _, command := range commands {
		if _, exists := byScript[command.Script]; exists {
			continue
		}
		candidates = append(candidates, command.Script)
		byScript[command.Script] = command
	}

	selected, err := FzfRunner(candidates)
	switch {
	case errors.Is(err, errFzfAborted):
		logs.Failed(newline + "Aborted")
		return nil, true
	case err != nil:
		logs.Debug("fzf selection failed, falling back to the built-in prompt: %v", err)
		return nil, false
	}

	command, ok := byScript[selected]
	if !ok {
		logs.Debug("fzf returned an unknown command: %q", selected)
		return nil, false
	}
	return command, true
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
