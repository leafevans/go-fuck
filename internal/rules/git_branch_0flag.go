package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/internal/shells"
	"github.com/leafevans/go-fuck/internal/specific"
	"github.com/leafevans/go-fuck/internal/types"
)

// gitBranch0flagFirst returns the first two character argument that starts
// with "0", or an empty string.
func gitBranch0flagFirst(parts []string) string {
	for _, part := range parts {
		if len(part) == 2 && strings.HasPrefix(part, "0") {
			return part
		}
	}
	return ""
}

// gitBranch0flagMatch reports whether a branch flag was mistyped with a zero.
func gitBranch0flagMatch(command *types.Command) bool {
	parts := command.ScriptParts()
	return len(parts) > 1 && parts[1] == "branch" && gitBranch0flagFirst(parts) != ""
}

// gitBranch0flagNewCommand replaces the zero with a dash. When the branch
// already exists, the old branch is deleted first.
func gitBranch0flagNewCommand(command *types.Command) ([]string, error) {
	branchName := gitBranch0flagFirst(command.ScriptParts())
	if branchName == "" {
		return nil, nil
	}

	fixedFlag := strings.ReplaceAll(branchName, "0", "-")
	fixedScript := strings.ReplaceAll(command.Script, branchName, fixedFlag)

	if strings.Contains(command.Output, "A branch named '") &&
		strings.Contains(command.Output, "' already exists.") {
		deleteBranch := "git branch -D " + branchName
		return []string{shells.Current().And(deleteBranch, fixedScript)}, nil
	}
	return []string{fixedScript}, nil
}

func init() {
	register(types.NewRule("git_branch_0flag",
		specific.GitSupportMatch(gitBranch0flagMatch),
		specific.GitSupportNew(gitBranch0flagNewCommand)))
}
