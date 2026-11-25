package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/v4/internal/specific"
	"github.com/leafevans/go-fuck/v4/internal/types"
	"github.com/leafevans/go-fuck/v4/internal/utils"
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
	return gitHasSubcommand(command.Script, "stash") && strings.Contains(command.Output, "usage:")
}

// gitFixStashNewCommand replaces the subcommand with its closest match or
// inserts the save subcommand.
func gitFixStashNewCommand(command *types.Command) ([]string, error) {
	_, stashCommand, ok := gitOperationWordSpan(command.Script, "stash")
	if !ok {
		return nil, nil
	}
	fixed := utils.GetClosest(stashCommand, gitFixStashCommands, 0.6, false)
	if fixed != "" {
		corrected, ok := gitReplaceOperation(command.Script, "stash", stashCommand, fixed)
		if !ok {
			return nil, nil
		}
		return []string{corrected}, nil
	}
	corrected, ok := gitInsertAfterSubcommand(command.Script, "stash", "save")
	if !ok {
		return nil, nil
	}
	return []string{corrected}, nil
}

func init() {
	register(types.NewRule("git_fix_stash",
		specific.GitSupportMatch(gitFixStashMatch),
		specific.GitSupportNew(gitFixStashNewCommand)))
}
