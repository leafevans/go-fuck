package rules

import (
	"github.com/leafevans/go-fuck/internal/specific"
	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// gitCommitAmendMatch reports whether a commit command was run.
func gitCommitAmendMatch(command *types.Command) bool {
	return utils.Contains(command.ScriptParts(), "commit")
}

// gitCommitAmendNewCommand amends the previous commit.
func gitCommitAmendNewCommand(command *types.Command) ([]string, error) {
	return []string{"git commit --amend"}, nil
}

func init() {
	register(types.NewRule("git_commit_amend",
		specific.GitSupportMatch(gitCommitAmendMatch),
		specific.GitSupportNew(gitCommitAmendNewCommand)))
}
