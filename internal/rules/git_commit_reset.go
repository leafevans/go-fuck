package rules

import (
	"github.com/leafevans/go-fuck/v2/internal/specific"
	"github.com/leafevans/go-fuck/v2/internal/types"
	"github.com/leafevans/go-fuck/v2/internal/utils"
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
	rule := types.NewRule("git_commit_reset",
		specific.GitSupportMatch(gitCommitResetMatch),
		specific.GitSupportNew(gitCommitResetNewCommand))
	// Deviation: the reference implementation gives this rule the default
	// priority, so it is suggested before any concrete fix. Resetting the last
	// commit is only a guess when a commit fails and is destructive; concrete
	// rules must be offered first.
	rule.Priority = 1500
	register(rule)
}
