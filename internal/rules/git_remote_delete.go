package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/internal/specific"
	"github.com/leafevans/go-fuck/internal/types"
)

// gitRemoteDeleteMatch reports whether the removed git subcommand is not a
// real git command.
func gitRemoteDeleteMatch(command *types.Command) bool {
	return strings.Contains(command.Script, "remote delete")
}

// gitRemoteDeleteNewCommand replaces delete with remove.
func gitRemoteDeleteNewCommand(command *types.Command) ([]string, error) {
	return []string{strings.Replace(command.Script, "delete", "remove", 1)}, nil
}

func init() {
	register(types.NewRule("git_remote_delete",
		specific.GitSupportMatch(gitRemoteDeleteMatch),
		specific.GitSupportNew(gitRemoteDeleteNewCommand)))
}
