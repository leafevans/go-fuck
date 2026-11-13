package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/v2/internal/shells"
	"github.com/leafevans/go-fuck/v2/internal/specific"
	"github.com/leafevans/go-fuck/v2/internal/types"
	"github.com/leafevans/go-fuck/v2/internal/utils"
)

// gitMergeUnstagedError matches the error that git prints when a merge would
// overwrite local changes.
const gitMergeUnstagedError = "would be overwritten by merge"

// gitMergeUnstagedMatch reports whether git merge failed because of local
// changes.
func gitMergeUnstagedMatch(command *types.Command) bool {
	return utils.Contains(command.ScriptParts(), "merge") &&
		strings.Contains(command.Output, gitMergeUnstagedError)
}

// gitMergeUnstagedNewCommand stashes the changes, merges and restores them.
func gitMergeUnstagedNewCommand(command *types.Command) ([]string, error) {
	return []string{shells.Current().And("git stash", command.Script, "git stash pop")}, nil
}

func init() {
	rule := types.NewRule("git_merge_unstaged",
		specific.GitSupportMatch(gitMergeUnstagedMatch),
		specific.GitSupportNew(gitMergeUnstagedNewCommand))
	// Deviation: git_stash of the reference implementation matches the same
	// output and suggests a stash without restoring the changes afterwards.
	// This rule ranks first so that the complete correction is suggested above
	// the plain one.
	rule.Priority = 900
	register(rule)
}
