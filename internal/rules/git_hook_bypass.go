package rules

import (
	"github.com/leafevans/go-fuck/internal/specific"
	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// gitHookBypassCommands lists the commands that run git hooks.
var gitHookBypassCommands = []string{"am", "commit", "push"}

// gitHookBypassMatch reports whether a command that runs git hooks is used.
func gitHookBypassMatch(command *types.Command) bool {
	parts := command.ScriptParts()
	for _, hookedCommand := range gitHookBypassCommands {
		if utils.Contains(parts, hookedCommand) {
			return true
		}
	}
	return false
}

// gitHookBypassNewCommand adds --no-verify to the hooked command.
func gitHookBypassNewCommand(command *types.Command) ([]string, error) {
	parts := command.ScriptParts()
	for _, hookedCommand := range gitHookBypassCommands {
		if utils.Contains(parts, hookedCommand) {
			return []string{utils.ReplaceArgument(
				command.Script, hookedCommand, hookedCommand+" --no-verify")}, nil
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
	register(rule)
}
