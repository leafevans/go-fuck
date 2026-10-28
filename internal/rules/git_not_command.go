package rules

import (
	"regexp"
	"strings"

	"github.com/leafevans/go-fuck/internal/specific"
	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// gitNotCommandBroken matches the command that is not a git command.
var gitNotCommandBroken = regexp.MustCompile("git: '([^']*)' is not a git command")

// gitNotCommandMatch reports whether git suggested similar commands.
func gitNotCommandMatch(command *types.Command) bool {
	return strings.Contains(command.Output, " is not a git command. See 'git --help'.") &&
		(strings.Contains(command.Output, "The most similar command") ||
			strings.Contains(command.Output, "Did you mean"))
}

// gitNotCommandNewCommand replaces the unknown command with the suggestions.
func gitNotCommandNewCommand(command *types.Command) ([]string, error) {
	brokenMatch := gitNotCommandBroken.FindStringSubmatch(command.Output)
	if brokenMatch == nil {
		return nil, nil
	}
	matched := utils.GetAllMatchedCommands(command.Output,
		"The most similar command", "Did you mean")
	return utils.ReplaceCommand(command, brokenMatch[1], matched), nil
}

func init() {
	register(types.NewRule("git_not_command",
		specific.GitSupportMatch(gitNotCommandMatch),
		specific.GitSupportNew(gitNotCommandNewCommand)))
}
