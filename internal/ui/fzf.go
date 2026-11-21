package ui

import (
	"errors"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/leafevans/go-fuck/v2/internal/logs"
	"github.com/leafevans/go-fuck/v2/internal/types"
	"github.com/leafevans/go-fuck/v2/internal/utils"
)

// errFzfUnavailable is returned when the fzf executable is not installed.
var errFzfUnavailable = errors.New("fzf is not available")

// errFzfAborted is returned when the user dismisses the fzf prompt.
var errFzfAborted = errors.New("fzf selection aborted")

// FzfRunner runs fzf with single-line display labels and returns the selected
// label. It is a variable so that tests can stub the fuzzy finder.
var FzfRunner = runFzf

// runFzf starts fzf with the candidate labels on standard input.
func runFzf(candidates []string) (string, error) {
	fzf := utils.Which("fzf")
	if fzf == "" {
		return "", errFzfUnavailable
	}

	args := []string{
		"--height=40%",
		"--reverse",
		"--no-multi",
		"--prompt=› ",
		"--header=enter: select   esc: abort",
	}
	if logs.ColorsEnabled() {
		args = append(args, "--color=prompt:214,pointer:214,header:245,info:245")
	} else {
		args = append(args, "--no-color")
	}
	command := exec.Command(fzf, args...)
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
	return strings.TrimSuffix(strings.TrimSuffix(string(output), newline), "\r"), nil
}

// fzfCandidates keeps display labels separate from executable scripts. Escaped
// controls and side-effect markers can otherwise collide with literal text.
func fzfCandidates(commands []*types.CorrectedCommand) ([]string, map[string]*types.CorrectedCommand) {
	candidates := make([]string, 0, len(commands))
	byLabel := make(map[string]*types.CorrectedCommand, len(commands))
	seen := types.NewCorrectedCommandSet(len(commands))
	for _, command := range commands {
		if !seen.Add(command) {
			continue
		}
		base := logs.CommandText(command.Script)
		if command.SideEffect != nil {
			base += "  (+side effect)"
		}
		label := base
		for number := 2; byLabel[label] != nil; number++ {
			label = base + "  [" + strconv.Itoa(number) + "]"
		}
		candidates = append(candidates, label)
		byLabel[label] = command
	}
	return candidates, byLabel
}

// selectWithFzf asks fzf to select one of the corrected commands. The second
// result is false when the caller should fall back to the built-in prompt.
func selectWithFzf(commands []*types.CorrectedCommand) (*types.CorrectedCommand, bool) {
	candidates, byLabel := fzfCandidates(commands)
	selected, err := FzfRunner(candidates)
	switch {
	case errors.Is(err, errFzfAborted):
		logs.NewLine()
		logs.Failed("Aborted")
		return nil, true
	case err != nil:
		logs.Debug("fzf selection failed, falling back to the built-in prompt: %v", err)
		return nil, false
	}

	command, ok := byLabel[selected]
	if !ok {
		logs.Debug("fzf returned an unknown command: %q", selected)
		return nil, false
	}
	return command, true
}
