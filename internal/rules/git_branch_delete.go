package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/internal/specific"
	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// gitBranchDeleteMatch reports whether git refused to delete an unmerged
// branch.
func gitBranchDeleteMatch(command *types.Command) bool {
	return strings.Contains(command.Script, "branch -d") &&
		strings.Contains(command.Output, "If you are sure you want to delete it")
}

// gitBranchDeleteNewCommand forces the branch deletion.
func gitBranchDeleteNewCommand(command *types.Command) ([]string, error) {
	return []string{utils.ReplaceArgument(command.Script, "-d", "-D")}, nil
}

func init() {
	register(types.NewRule("git_branch_delete",
		specific.GitSupportMatch(gitBranchDeleteMatch),
		specific.GitSupportNew(gitBranchDeleteNewCommand)))
}
