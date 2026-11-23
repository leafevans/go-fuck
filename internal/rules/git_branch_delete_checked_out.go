package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/v3/internal/shells"
	"github.com/leafevans/go-fuck/v3/internal/specific"
	"github.com/leafevans/go-fuck/v3/internal/types"
)

// gitBranchDeleteCheckedOutMatch reports whether git refused to delete the
// currently checked out branch.
func gitBranchDeleteCheckedOutMatch(command *types.Command) bool {
	_, _, lower := specific.GitOptionSpan(command.Script, "-d", shells.Current())
	_, _, upper := specific.GitOptionSpan(command.Script, "-D", shells.Current())
	return gitHasSubcommand(command.Script, "branch") && (lower || upper) &&
		strings.Contains(command.Output, "error: Cannot delete branch '") &&
		strings.Contains(command.Output, "' checked out at '")
}

// gitBranchDeleteCheckedOutNewCommand checks out master and force deletes the
// branch.
func gitBranchDeleteCheckedOutNewCommand(command *types.Command) ([]string, error) {
	corrected := command.Script
	if _, _, ok := specific.GitOptionSpan(command.Script, "-d", shells.Current()); ok {
		var replaced bool
		corrected, replaced = gitReplaceArgument(command.Script, "-d", "-D", false)
		if !replaced {
			return nil, nil
		}
	} else if _, _, ok := specific.GitOptionSpan(command.Script, "-D", shells.Current()); !ok {
		return nil, nil
	}
	checkout, ok := gitCommandWithSubcommand(command.Script, "branch", "checkout master")
	if !ok {
		return nil, nil
	}
	return []string{shells.Current().And(
		checkout,
		corrected)}, nil
}

func init() {
	register(types.NewRule("git_branch_delete_checked_out",
		specific.GitSupportMatch(gitBranchDeleteCheckedOutMatch),
		specific.GitSupportNew(gitBranchDeleteCheckedOutNewCommand)))
}
