package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/internal/specific"
	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// gitTagForceMatch reports whether git tag failed because the tag exists.
func gitTagForceMatch(command *types.Command) bool {
	return utils.Contains(command.ScriptParts(), "tag") &&
		strings.Contains(command.Output, "already exists")
}

// gitTagForceNewCommand forces the tag creation.
func gitTagForceNewCommand(command *types.Command) ([]string, error) {
	return []string{utils.ReplaceArgument(command.Script, "tag", "tag --force")}, nil
}

func init() {
	register(types.NewRule("git_tag_force",
		specific.GitSupportMatch(gitTagForceMatch),
		specific.GitSupportNew(gitTagForceNewCommand)))
}
