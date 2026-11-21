package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/v2/internal/shells"
	"github.com/leafevans/go-fuck/v2/internal/specific"
	"github.com/leafevans/go-fuck/v2/internal/types"
)

// gitBranchDeleteMatch reports whether git refused to delete an unmerged
// branch.
func gitBranchDeleteMatch(command *types.Command) bool {
	_, _, ok := specific.GitOptionSpan(command.Script, "-d", shells.Current())
	return gitHasSubcommand(command.Script, "branch") && ok &&
		strings.Contains(command.Output, "If you are sure you want to delete it")
}

// gitBranchDeleteNewCommand forces the branch deletion.
func gitBranchDeleteNewCommand(command *types.Command) ([]string, error) {
	corrected, ok := gitReplaceArgument(command.Script, "-d", "-D", false)
	if !ok {
		return nil, nil
	}
	return []string{corrected}, nil
}

func init() {
	register(types.NewRule("git_branch_delete",
		specific.GitSupportMatch(gitBranchDeleteMatch),
		specific.GitSupportNew(gitBranchDeleteNewCommand)))
}
