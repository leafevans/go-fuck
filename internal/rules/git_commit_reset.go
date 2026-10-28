package rules

import (
	"github.com/leafevans/go-fuck/internal/specific"
	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// gitCommitResetMatch reports whether a commit command was run.
func gitCommitResetMatch(command *types.Command) bool {
	return utils.Contains(command.ScriptParts(), "commit")
}

// gitCommitResetNewCommand resets the last commit.
func gitCommitResetNewCommand(command *types.Command) ([]string, error) {
	return []string{"git reset HEAD~"}, nil
}

func init() {
	register(types.NewRule("git_commit_reset",
		specific.GitSupportMatch(gitCommitResetMatch),
		specific.GitSupportNew(gitCommitResetNewCommand)))
}
