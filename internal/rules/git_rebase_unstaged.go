package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/v2/internal/shells"
	"github.com/leafevans/go-fuck/v2/internal/specific"
	"github.com/leafevans/go-fuck/v2/internal/types"
	"github.com/leafevans/go-fuck/v2/internal/utils"
)

// gitRebaseUnstagedError matches the error that git prints when a rebase is
// blocked by local changes.
const gitRebaseUnstagedError = "cannot rebase"

// gitRebaseUnstagedMatch reports whether git rebase failed because of local
// changes.
func gitRebaseUnstagedMatch(command *types.Command) bool {
	if !utils.Contains(command.ScriptParts(), "rebase") {
		return false
	}
	output := command.Output
	return strings.Contains(output, gitRebaseUnstagedError) &&
		(strings.Contains(output, "unstaged changes") ||
			strings.Contains(output, "uncommitted changes"))
}

// gitRebaseUnstagedNewCommand stashes the changes, rebases and restores them.
func gitRebaseUnstagedNewCommand(command *types.Command) ([]string, error) {
	return []string{shells.Current().And("git stash", command.Script, "git stash pop")}, nil
}

func init() {
	rule := types.NewRule("git_rebase_unstaged",
		specific.GitSupportMatch(gitRebaseUnstagedMatch),
		specific.GitSupportNew(gitRebaseUnstagedNewCommand))
	// Deviation: git_stash of the reference implementation matches the same
	// output and suggests a stash without restoring the changes afterwards.
	// This rule ranks first so that the complete correction is suggested above
	// the plain one.
	rule.Priority = 900
	register(rule)
}
