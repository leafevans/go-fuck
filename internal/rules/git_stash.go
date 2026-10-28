package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/internal/shells"
	"github.com/leafevans/go-fuck/internal/specific"
	"github.com/leafevans/go-fuck/internal/types"
)

// gitStashMatch reports whether git asked to commit or stash the local
// changes.
func gitStashMatch(command *types.Command) bool {
	// Catches "Please commit or stash them" and "Please, commit your changes
	// or stash them before you can switch branches.".
	return strings.Contains(command.Output, "or stash them")
}

// gitStashNewCommand stashes the changes and retries the command.
func gitStashNewCommand(command *types.Command) ([]string, error) {
	return []string{shells.Current().And("git stash", command.Script)}, nil
}

func init() {
	register(types.NewRule("git_stash",
		specific.GitSupportMatch(gitStashMatch),
		specific.GitSupportNew(gitStashNewCommand)))
}
