package rules

import (
	"regexp"

	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// unknownCommandBrokenCommand matches the unknown command reported by the
// application.
var unknownCommandBrokenCommand = regexp.MustCompile(`([^:]*): Unknown command.*`)

// unknownCommandSuggestion matches the suggestions printed by the application.
var unknownCommandSuggestion = regexp.MustCompile(`Did you mean ([^?]*)?`)

// unknownCommandMatch reports whether the application printed a list of similar
// commands.
func unknownCommandMatch(command *types.Command) bool {
	return unknownCommandBrokenCommand.MatchString(command.Output) &&
		unknownCommandSuggestion.MatchString(command.Output)
}

// unknownCommandNewCommand replaces the unknown command with the suggested one.
func unknownCommandNewCommand(command *types.Command) ([]string, error) {
	broken := unknownCommandBrokenCommand.FindAllStringSubmatch(command.Output, -1)
	if len(broken) == 0 {
		return nil, nil
	}

	suggestions := []string{}
	for _, match := range unknownCommandSuggestion.FindAllStringSubmatch(command.Output, -1) {
		suggestions = append(suggestions, match[1])
	}
	return utils.ReplaceCommand(command, broken[0][1], suggestions), nil
}

func init() {
	register(types.NewRule("unknown_command", unknownCommandMatch, unknownCommandNewCommand))
}
