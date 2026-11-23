package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/v3/internal/specific"
	"github.com/leafevans/go-fuck/v3/internal/types"
)

// gitAddForceMatch reports whether git refused to add ignored files.
func gitAddForceMatch(command *types.Command) bool {
	return gitHasSubcommand(command.Script, "add") &&
		strings.Contains(command.Output, "Use -f if you really want to add them.")
}

// gitAddForceNewCommand adds the --force flag to git add.
func gitAddForceNewCommand(command *types.Command) ([]string, error) {
	corrected, ok := gitInsertAfterSubcommand(command.Script, "add", "--force")
	if !ok {
		return nil, nil
	}
	return []string{corrected}, nil
}

func init() {
	register(types.NewRule("git_add_force",
		specific.GitSupportMatch(gitAddForceMatch),
		specific.GitSupportNew(gitAddForceNewCommand)))
}
