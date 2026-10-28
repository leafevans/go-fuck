package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/internal/specific"
	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// gitPullCloneMatch reports whether git pull was run outside of a repository.
func gitPullCloneMatch(command *types.Command) bool {
	return strings.Contains(command.Output, "fatal: Not a git repository") &&
		strings.Contains(command.Output,
			"Stopping at filesystem boundary (GIT_DISCOVERY_ACROSS_FILESYSTEM not set).")
}

// gitPullCloneNewCommand clones the repository instead of pulling it.
func gitPullCloneNewCommand(command *types.Command) ([]string, error) {
	return []string{utils.ReplaceArgument(command.Script, "pull", "clone")}, nil
}

func init() {
	register(types.NewRule("git_pull_clone",
		specific.GitSupportMatch(gitPullCloneMatch),
		specific.GitSupportNew(gitPullCloneNewCommand)))
}
