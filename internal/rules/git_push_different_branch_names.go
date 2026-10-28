package rules

import (
	"regexp"
	"strings"

	"github.com/leafevans/go-fuck/internal/specific"
	"github.com/leafevans/go-fuck/internal/types"
)

// gitPushDifferentBranchSuggestion matches the commands git suggests when the
// upstream branch name differs from the local one.
var gitPushDifferentBranchSuggestion = regexp.MustCompile(`(?m)^ +(git push \S+ \S+)`)

// gitPushDifferentBranchNamesMatch reports whether the upstream branch has a
// different name than the current branch.
func gitPushDifferentBranchNamesMatch(command *types.Command) bool {
	return strings.Contains(command.Script, "push") &&
		strings.Contains(command.Output,
			"The upstream branch of your current branch does not match")
}

// gitPushDifferentBranchNamesNewCommand uses the first command suggested by
// git.
func gitPushDifferentBranchNamesNewCommand(command *types.Command) ([]string, error) {
	matches := gitPushDifferentBranchSuggestion.FindAllStringSubmatch(command.Output, -1)
	if len(matches) == 0 {
		return nil, nil
	}
	return []string{matches[0][1]}, nil
}

func init() {
	register(types.NewRule("git_push_different_branch_names",
		specific.GitSupportMatch(gitPushDifferentBranchNamesMatch),
		specific.GitSupportNew(gitPushDifferentBranchNamesNewCommand)))
}
