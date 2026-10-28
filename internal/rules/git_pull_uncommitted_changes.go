package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/internal/shells"
	"github.com/leafevans/go-fuck/internal/specific"
	"github.com/leafevans/go-fuck/internal/types"
)

// gitPullUncommittedChangesMatch reports whether git pull failed because of
// local changes.
func gitPullUncommittedChangesMatch(command *types.Command) bool {
	return strings.Contains(command.Script, "pull") &&
		(strings.Contains(command.Output, "You have unstaged changes") ||
			strings.Contains(command.Output, "contains uncommitted changes"))
}

// gitPullUncommittedChangesNewCommand stashes the local changes, pulls and
// restores them.
func gitPullUncommittedChangesNewCommand(command *types.Command) ([]string, error) {
	return []string{shells.Current().And("git stash", "git pull", "git stash pop")}, nil
}

func init() {
	register(types.NewRule("git_pull_uncommitted_changes",
		specific.GitSupportMatch(gitPullUncommittedChangesMatch),
		specific.GitSupportNew(gitPullUncommittedChangesNewCommand)))
}
