package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/internal/specific"
	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// gitAddForceMatch reports whether git refused to add ignored files.
func gitAddForceMatch(command *types.Command) bool {
	return utils.Contains(command.ScriptParts(), "add") &&
		strings.Contains(command.Output, "Use -f if you really want to add them.")
}

// gitAddForceNewCommand adds the --force flag to git add.
func gitAddForceNewCommand(command *types.Command) ([]string, error) {
	return []string{utils.ReplaceArgument(command.Script, "add", "add --force")}, nil
}

func init() {
	register(types.NewRule("git_add_force",
		specific.GitSupportMatch(gitAddForceMatch),
		specific.GitSupportNew(gitAddForceNewCommand)))
}
