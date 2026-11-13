package rules

import (
	"github.com/leafevans/go-fuck/v2/internal/shells"
	"github.com/leafevans/go-fuck/v2/internal/specific"
	"github.com/leafevans/go-fuck/v2/internal/types"
)

// gitBranchListMatch reports whether "git branch list" was used instead of
// "git branch".
func gitBranchListMatch(command *types.Command) bool {
	parts := command.ScriptParts()
	return len(parts) == 3 && parts[1] == "branch" && parts[2] == "list"
}

// gitBranchListNewCommand deletes the branch called "list" and lists the
// remaining branches.
func gitBranchListNewCommand(command *types.Command) ([]string, error) {
	return []string{shells.Current().And("git branch --delete list", "git branch")}, nil
}

func init() {
	register(types.NewRule("git_branch_list",
		specific.GitSupportMatch(gitBranchListMatch),
		specific.GitSupportNew(gitBranchListNewCommand)))
}
