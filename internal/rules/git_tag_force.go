package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/v3/internal/specific"
	"github.com/leafevans/go-fuck/v3/internal/types"
)

// gitTagForceMatch reports whether git tag failed because the tag exists.
func gitTagForceMatch(command *types.Command) bool {
	return gitHasSubcommand(command.Script, "tag") &&
		strings.Contains(command.Output, "already exists")
}

// gitTagForceNewCommand forces the tag creation.
func gitTagForceNewCommand(command *types.Command) ([]string, error) {
	corrected, ok := gitInsertAfterSubcommand(command.Script, "tag", "--force")
	if !ok {
		return nil, nil
	}
	return []string{corrected}, nil
}

func init() {
	register(types.NewRule("git_tag_force",
		specific.GitSupportMatch(gitTagForceMatch),
		specific.GitSupportNew(gitTagForceNewCommand)))
}
