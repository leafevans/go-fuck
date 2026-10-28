package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/internal/specific"
	"github.com/leafevans/go-fuck/internal/types"
)

// gitCloneGitCloneMatch reports whether "git clone" was passed to "git clone"
// twice.
func gitCloneGitCloneMatch(command *types.Command) bool {
	return strings.Contains(command.Script, " git clone ") &&
		strings.Contains(command.Output, "fatal: Too many arguments.")
}

// gitCloneGitCloneNewCommand removes the duplicated "git clone".
func gitCloneGitCloneNewCommand(command *types.Command) ([]string, error) {
	return []string{strings.Replace(command.Script, " git clone ", " ", 1)}, nil
}

func init() {
	register(types.NewRule("git_clone_git_clone",
		specific.GitSupportMatch(gitCloneGitCloneMatch),
		specific.GitSupportNew(gitCloneGitCloneNewCommand)))
}
