package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/internal/specific"
	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// gitDiffStagedMatch reports whether git diff was called without --staged.
func gitDiffStagedMatch(command *types.Command) bool {
	return strings.Contains(command.Script, "diff") &&
		!strings.Contains(command.Script, "--staged")
}

// gitDiffStagedNewCommand adds the --staged flag.
func gitDiffStagedNewCommand(command *types.Command) ([]string, error) {
	return []string{utils.ReplaceArgument(command.Script, "diff", "diff --staged")}, nil
}

func init() {
	register(types.NewRule("git_diff_staged",
		specific.GitSupportMatch(gitDiffStagedMatch),
		specific.GitSupportNew(gitDiffStagedNewCommand)))
}
