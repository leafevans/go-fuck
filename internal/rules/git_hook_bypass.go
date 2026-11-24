package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/v3/internal/specific"
	"github.com/leafevans/go-fuck/v3/internal/types"
)

// gitHookBypassCommands lists the commands that run git hooks.
var gitHookBypassCommands = []string{"am", "commit", "push"}

// gitHookFailureMarkers list the output markers of a git hook that rejected the
// command.
var gitHookFailureMarkers = []string{"hook", "husky", "lint-staged"}

// gitHookFailure reports whether the command output names a git hook.
func gitHookFailure(output string) bool {
	lower := strings.ToLower(output)
	for _, marker := range gitHookFailureMarkers {
		if strings.Contains(lower, marker) {
			return true
		}
	}
	return false
}

// gitHookBypassMatch reports whether a git hook rejected a command that runs
// hooks.
//
// Deviation: the reference implementation matches the command alone, so every
// failed "git commit", "git am" or "git push" was answered with --no-verify,
// including failures a hook never caused, such as a rejected non-fast-forward
// push or a missing commit identity. The output has to name the hook, and a
// command that already passes --no-verify is left alone.
func gitHookBypassMatch(command *types.Command) bool {
	if strings.Contains(command.Script, "--no-verify") || !gitHookFailure(command.Output) {
		return false
	}
	for _, hookedCommand := range gitHookBypassCommands {
		if gitHasSubcommand(command.Script, hookedCommand) {
			return true
		}
	}
	return false
}

// gitHookBypassNewCommand adds --no-verify to the hooked command.
func gitHookBypassNewCommand(command *types.Command) ([]string, error) {
	for _, hookedCommand := range gitHookBypassCommands {
		if corrected, ok := gitInsertAfterSubcommand(command.Script, hookedCommand, "--no-verify"); ok {
			return []string{corrected}, nil
		}
	}
	return nil, nil
}

func init() {
	rule := types.NewRule("git_hook_bypass",
		specific.GitSupportMatch(gitHookBypassMatch),
		specific.GitSupportNew(gitHookBypassNewCommand))
	rule.Priority = 1100
	rule.Reason = "Add --no-verify to bypass the git hook that rejected the command"
	register(rule)
}
