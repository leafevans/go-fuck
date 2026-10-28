package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/internal/shells"
	"github.com/leafevans/go-fuck/internal/specific"
	"github.com/leafevans/go-fuck/internal/types"
)

// gitStashPopMatch reports whether git stash pop failed because local changes
// would be overwritten.
func gitStashPopMatch(command *types.Command) bool {
	return strings.Contains(command.Script, "stash") &&
		strings.Contains(command.Script, "pop") &&
		strings.Contains(command.Output,
			"Your local changes to the following files would be overwritten by merge")
}

// gitStashPopNewCommand stages the local changes before applying the stash.
func gitStashPopNewCommand(command *types.Command) ([]string, error) {
	return []string{shells.Current().And(
		"git add --update", "git stash pop", "git reset .")}, nil
}

func init() {
	rule := types.NewRule("git_stash_pop",
		specific.GitSupportMatch(gitStashPopMatch),
		specific.GitSupportNew(gitStashPopNewCommand))
	// Python: priority = 900, so that it comes before the other applicable
	// rules.
	rule.Priority = 900
	register(rule)
}
