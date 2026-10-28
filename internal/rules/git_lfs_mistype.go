package rules

import (
	"regexp"
	"strings"

	"github.com/leafevans/go-fuck/internal/specific"
	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// gitLfsMistypeBroken matches the mistyped git-lfs command.
var gitLfsMistypeBroken = regexp.MustCompile(
	"Error: unknown command \"([^\"]*)\" for \"git-lfs\"")

// gitLfsMistypeMatch reports whether git-lfs printed the usage of a mistyped
// command.
func gitLfsMistypeMatch(command *types.Command) bool {
	return strings.Contains(command.Script, "lfs") &&
		strings.Contains(command.Output, "Did you mean this?")
}

// gitLfsMistypeNewCommand replaces the subcommand with its closest matches.
func gitLfsMistypeNewCommand(command *types.Command) ([]string, error) {
	brokenMatch := gitLfsMistypeBroken.FindStringSubmatch(command.Output)
	if brokenMatch == nil {
		return nil, nil
	}
	matched := utils.GetAllMatchedCommands(command.Output, "Did you mean", " for usage.")
	return utils.ReplaceCommand(command, brokenMatch[1], matched), nil
}

func init() {
	register(types.NewRule("git_lfs_mistype",
		specific.GitSupportMatch(gitLfsMistypeMatch),
		specific.GitSupportNew(gitLfsMistypeNewCommand)))
}
