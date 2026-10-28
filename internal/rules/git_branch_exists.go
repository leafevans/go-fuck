package rules

import (
	"regexp"
	"strings"

	"github.com/leafevans/go-fuck/internal/shells"
	"github.com/leafevans/go-fuck/internal/specific"
	"github.com/leafevans/go-fuck/internal/types"
)

// gitBranchExistsBranch matches the branch name reported by git.
var gitBranchExistsBranch = regexp.MustCompile("fatal: A branch named '(.+)' already exists.")

// gitBranchExistsMatch reports whether the branch that should be created
// already exists.
func gitBranchExistsMatch(command *types.Command) bool {
	return strings.Contains(command.Output, "fatal: A branch named '") &&
		strings.Contains(command.Output, "' already exists.")
}

// gitBranchExistsNewCommand suggests removing or reusing the existing branch.
func gitBranchExistsNewCommand(command *types.Command) ([]string, error) {
	branchMatch := gitBranchExistsBranch.FindStringSubmatch(command.Output)
	if branchMatch == nil {
		return nil, nil
	}
	branchName := strings.ReplaceAll(branchMatch[1], "'", backslash+"'")

	templates := [][]string{
		{"git branch -d {0}", "git branch {0}"},
		{"git branch -d {0}", "git checkout -b {0}"},
		{"git branch -D {0}", "git branch {0}"},
		{"git branch -D {0}", "git checkout -b {0}"},
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
