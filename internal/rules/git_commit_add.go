package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/internal/specific"
	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// gitCommitAddMatch reports whether git complained that nothing was staged.
func gitCommitAddMatch(command *types.Command) bool {
	return utils.Contains(command.ScriptParts(), "commit") &&
		strings.Contains(command.Output, "no changes added to commit")
}

// gitCommitAddNewCommand stages changes before committing again.
func gitCommitAddNewCommand(command *types.Command) ([]string, error) {
	return []string{
		utils.ReplaceArgument(command.Script, "commit", "commit -a"),
		utils.ReplaceArgument(command.Script, "commit", "commit -p"),
	}, nil
}

func init() {
	register(types.NewRule("git_commit_add",
		specific.GitSupportMatch(gitCommitAddMatch),
		specific.GitSupportNew(gitCommitAddNewCommand)))
}
