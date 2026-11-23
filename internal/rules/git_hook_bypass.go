package rules

import (
	"github.com/leafevans/go-fuck/v3/internal/specific"
	"github.com/leafevans/go-fuck/v3/internal/types"
)

// gitHookBypassCommands lists the commands that run git hooks.
var gitHookBypassCommands = []string{"am", "commit", "push"}

// gitHookBypassMatch reports whether a command that runs git hooks is used.
func gitHookBypassMatch(command *types.Command) bool {
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
	rule.RequiresOutput = false
	rule.Reason = "Add --no-verify to bypass Git hooks that the flag disables"
	register(rule)
}
