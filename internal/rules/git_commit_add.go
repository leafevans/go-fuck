package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/v3/internal/specific"
	"github.com/leafevans/go-fuck/v3/internal/types"
)

// gitCommitAddMatch reports whether git complained that nothing was staged.
func gitCommitAddMatch(command *types.Command) bool {
	return gitHasSubcommand(command.Script, "commit") &&
		strings.Contains(command.Output, "no changes added to commit")
}

// gitCommitAddNewCommand stages changes before committing again.
func gitCommitAddNewCommand(command *types.Command) ([]string, error) {
	all, ok := gitInsertAfterSubcommand(command.Script, "commit", "-a")
	if !ok {
		return nil, nil
	}
	patch, _ := gitInsertAfterSubcommand(command.Script, "commit", "-p")
	return []string{all, patch}, nil
}

func init() {
	register(types.NewRule("git_commit_add",
		specific.GitSupportMatch(gitCommitAddMatch),
		specific.GitSupportNew(gitCommitAddNewCommand)))
}
