package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/v4/internal/shells"
	"github.com/leafevans/go-fuck/v4/internal/specific"
	"github.com/leafevans/go-fuck/v4/internal/types"
)

// gitBranchExistsBranch matches the branch name reported by git.
var gitBranchExistsBranch = newRegexp("fatal: A branch named '(.+)' already exists.")

// gitBranchExistsMatch reports whether the branch that should be created
// already exists.
func gitBranchExistsMatch(command *types.Command) bool {
	return strings.Contains(command.Output, "fatal: A branch named '") &&
		strings.Contains(command.Output, "' already exists.")
}

// gitBranchExistsNewCommand reuses the existing branch.
//
// Deviation: do not suggest deleting and recreating an existing branch, even
// with -d: recreating at HEAD can move its tip and -D can discard unmerged work.
// An existing branch is evidence for reuse, not consent to replace it.
func gitBranchExistsNewCommand(command *types.Command) ([]string, error) {
	branchMatch := gitBranchExistsBranch.FindStringSubmatch(command.Output)
	if branchMatch == nil {
		return nil, nil
	}
	branchName := strings.ReplaceAll(branchMatch[1], "'", backslash+"'")

	templates := [][]string{
		{"git checkout {0}"},
	}
	results := make([]string, 0, len(templates))
	for _, template := range templates {
		parts := make([]string, len(template))
		for index, part := range template {
			parts[index] = strings.ReplaceAll(part, "{0}", branchName)
		}
		results = append(results, shells.Current().And(parts...))
	}
	return results, nil
}

func init() {
	register(types.NewRule("git_branch_exists",
		specific.GitSupportMatch(gitBranchExistsMatch),
		specific.GitSupportNew(gitBranchExistsNewCommand)))
}
