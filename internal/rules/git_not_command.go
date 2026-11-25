package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/v4/internal/shells"
	"github.com/leafevans/go-fuck/v4/internal/specific"
	"github.com/leafevans/go-fuck/v4/internal/types"
	"github.com/leafevans/go-fuck/v4/internal/utils"
)

// gitNotCommandBroken matches the command that is not a git command.
var gitNotCommandBroken = newRegexp("git: '([^']*)' is not a git command")

var gitNotCommandCandidate = newRegexp(`^[A-Za-z0-9][A-Za-z0-9_.-]*$`)

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
	shell := shells.Current()
	broken := brokenMatch[1]
	start, end, ok := specific.GitSubcommandSpan(command.Script, broken, shell)
	if !ok {
		return nil, nil
	}
	matched := utils.GetAllMatchedCommands(command.Output,
		"The most similar command", "Did you mean")
	candidates := matched[:0]
	seen := make(map[string]bool, len(matched))
	for _, candidate := range matched {
		if candidate != broken && !seen[candidate] && gitNotCommandCandidate.MatchString(candidate) {
			seen[candidate] = true
			candidates = append(candidates, candidate)
		}
	}
	ranked := utils.GetCloseMatches(broken, candidates, 0, 0.1)
	for i, candidate := range ranked {
		ranked[i] = command.Script[:start] + candidate + command.Script[end:]
	}
	return ranked, nil
}

func init() {
	register(types.NewRule("git_not_command",
		specific.GitSupportMatch(gitNotCommandMatch),
		specific.GitSupportNew(gitNotCommandNewCommand)))
}
