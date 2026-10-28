package rules

import (
	"regexp"
	"strings"

	"github.com/leafevans/go-fuck/internal/types"
)

// removeShellPromptLiteralPattern matches a script that starts with a shell
// prompt and a command, as copied from a documentation code block.
var removeShellPromptLiteralPattern = regexp.MustCompile(`^\s*\$ \S+`)

// removeShellPromptLiteralMatch reports whether the failed command contains a
// literal shell prompt.
func removeShellPromptLiteralMatch(command *types.Command) bool {
	return strings.Contains(command.Output, "$: command not found") &&
		removeShellPromptLiteralPattern.MatchString(command.Script)
}

// removeShellPromptLiteralNewCommand removes the literal shell prompts.
func removeShellPromptLiteralNewCommand(command *types.Command) ([]string, error) {
	return []string{strings.TrimLeft(command.Script, "$ ")}, nil
}

func init() {
	register(types.NewRule("remove_shell_prompt_literal",
		removeShellPromptLiteralMatch,
		removeShellPromptLiteralNewCommand))
}
