package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/v2/internal/specific"
	"github.com/leafevans/go-fuck/v2/internal/types"
)

// gitPullCloneMatch reports whether git pull was run outside of a repository.
func gitPullCloneMatch(command *types.Command) bool {
	return gitHasSubcommand(command.Script, "pull") &&
		strings.Contains(command.Output, "fatal: Not a git repository") &&
		strings.Contains(command.Output,
			"Stopping at filesystem boundary (GIT_DISCOVERY_ACROSS_FILESYSTEM not set).")
}

// gitPullCloneNewCommand clones the repository instead of pulling it.
func gitPullCloneNewCommand(command *types.Command) ([]string, error) {
	corrected, ok := gitReplaceSubcommand(command.Script, "pull", "clone")
	if !ok {
		return nil, nil
	}
	return []string{corrected}, nil
}

func init() {
	register(types.NewRule("git_pull_clone",
		specific.GitSupportMatch(gitPullCloneMatch),
		specific.GitSupportNew(gitPullCloneNewCommand)))
}
