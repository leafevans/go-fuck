package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/internal/shells"
	"github.com/leafevans/go-fuck/internal/specific"
	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// gitBranchDeleteCheckedOutMatch reports whether git refused to delete the
// currently checked out branch.
func gitBranchDeleteCheckedOutMatch(command *types.Command) bool {
	return (strings.Contains(command.Script, "branch -d") ||
		strings.Contains(command.Script, "branch -D")) &&
		strings.Contains(command.Output, "error: Cannot delete branch '") &&
		strings.Contains(command.Output, "' checked out at '")
}

// gitBranchDeleteCheckedOutNewCommand checks out master and force deletes the
// branch.
func gitBranchDeleteCheckedOutNewCommand(command *types.Command) ([]string, error) {
	return []string{shells.Current().And(
		"git checkout master",
		utils.ReplaceArgument(command.Script, "-d", "-D"))}, nil
}

func init() {
	register(types.NewRule("git_branch_delete_checked_out",
		specific.GitSupportMatch(gitBranchDeleteCheckedOutMatch),
		specific.GitSupportNew(gitBranchDeleteCheckedOutNewCommand)))
}
