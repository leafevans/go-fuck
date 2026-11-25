package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/v4/internal/specific"
	"github.com/leafevans/go-fuck/v4/internal/types"
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

// gitBypassableHookFailure requires a local hook name and a failure diagnostic
// on the same line. Server-side hooks are not disabled by --no-verify.
func gitBypassableHookFailure(output string) bool {
	for _, line := range strings.Split(strings.ToLower(output), "\n") {
		if strings.Contains(line, "remote:") || strings.Contains(line, "pre-receive") || strings.Contains(line, "update hook") {
			continue
		}
		local := strings.Contains(line, "pre-commit") || strings.Contains(line, "commit-msg") || strings.Contains(line, "pre-push") || strings.Contains(line, "applypatch-msg") || strings.Contains(line, "pre-applypatch") || strings.HasPrefix(strings.TrimSpace(line), "husky -") || strings.HasPrefix(strings.TrimSpace(line), "lint-staged:")
		failed := strings.Contains(line, "failed") || strings.Contains(line, "exited with code") || strings.Contains(line, "exit code 1") || strings.Contains(line, "error:")
		if local && failed {
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
	if strings.Contains(command.Script, "--no-verify") || !gitBypassableHookFailure(command.Output) {
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
