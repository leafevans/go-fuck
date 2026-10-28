package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/internal/specific"
	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// gitFixStashCommands lists the valid git stash subcommands.
var gitFixStashCommands = []string{
	"apply",
	"branch",
	"clear",
	"drop",
	"list",
	"pop",
	"save",
	"show",
}

// gitFixStashMatch reports whether a mistyped git stash subcommand was used.
func gitFixStashMatch(command *types.Command) bool {
	parts := command.ScriptParts()
	return len(parts) > 1 && parts[1] == "stash" && strings.Contains(command.Output, "usage:")
}

// gitFixStashNewCommand replaces the subcommand with its closest match or
// inserts the save subcommand.
func gitFixStashNewCommand(command *types.Command) ([]string, error) {
	parts := command.ScriptParts()
	if len(parts) < 3 {
		return nil, nil
	}
	stashCommand := parts[2]
	fixed := utils.GetClosest(stashCommand, gitFixStashCommands, 0.6, false)
	if fixed != "" {
		return []string{utils.ReplaceArgument(command.Script, stashCommand, fixed)}, nil
	}

	fixedParts := make([]string, 0, len(parts)+1)
	fixedParts = append(fixedParts, parts[:2]...)
	fixedParts = append(fixedParts, "save")
	fixedParts = append(fixedParts, parts[2:]...)
	return []string{strings.Join(fixedParts, " ")}, nil
}

func init() {
	register(types.NewRule("git_fix_stash",
		specific.GitSupportMatch(gitFixStashMatch),
		specific.GitSupportNew(gitFixStashNewCommand)))
}
