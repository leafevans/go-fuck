package rules

import (
	"regexp"
	"strings"

	"github.com/leafevans/go-fuck/internal/specific"
	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// gitMergeUnknownBranch matches the branch that git could not merge.
var gitMergeUnknownBranch = regexp.MustCompile("merge: (.+) - not something we can merge")

// gitMergeRemoteBranch matches the suggestion printed by git.
var gitMergeRemoteBranch = regexp.MustCompile(`Did you mean this\?\n\t([^\n]+)`)

// gitMergeMatch reports whether git suggested a similar remote branch.
func gitMergeMatch(command *types.Command) bool {
	return strings.Contains(command.Script, "merge") &&
		strings.Contains(command.Output, " - not something we can merge") &&
		strings.Contains(command.Output, "Did you mean this?")
}

// gitMergeNewCommand replaces the unknown branch with the remote one.
func gitMergeNewCommand(command *types.Command) ([]string, error) {
	unknownMatch := gitMergeUnknownBranch.FindStringSubmatch(command.Output)
	remoteMatch := gitMergeRemoteBranch.FindStringSubmatch(command.Output)
	if unknownMatch == nil || remoteMatch == nil {
		return nil, nil
	}
	return []string{utils.ReplaceArgument(command.Script, unknownMatch[1], remoteMatch[1])}, nil
}

func init() {
	register(types.NewRule("git_merge",
		specific.GitSupportMatch(gitMergeMatch),
		specific.GitSupportNew(gitMergeNewCommand)))
}
