package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/internal/specific"
	"github.com/leafevans/go-fuck/internal/types"
)

// gitMergeUnrelatedMatch reports whether git refused to merge unrelated
// histories.
func gitMergeUnrelatedMatch(command *types.Command) bool {
	return strings.Contains(command.Script, "merge") &&
		strings.Contains(command.Output, "fatal: refusing to merge unrelated histories")
}

// gitMergeUnrelatedNewCommand allows merging unrelated histories.
func gitMergeUnrelatedNewCommand(command *types.Command) ([]string, error) {
	return []string{command.Script + " --allow-unrelated-histories"}, nil
}

func init() {
	register(types.NewRule("git_merge_unrelated",
		specific.GitSupportMatch(gitMergeUnrelatedMatch),
		specific.GitSupportNew(gitMergeUnrelatedNewCommand)))
}
