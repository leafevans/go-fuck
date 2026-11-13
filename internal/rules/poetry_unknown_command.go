package rules

import (
	"regexp"
	"strings"

	"github.com/leafevans/go-fuck/v2/internal/types"
	"github.com/leafevans/go-fuck/v2/internal/utils"
)

// poetryUnknownCommand matches the error line of Poetry for an unknown
// command, for example: The requested command intall does not exist.
var poetryUnknownCommand = regexp.MustCompile(`The requested command (\S+) does not exist\.`)

// poetrySuggestionHeader introduces the list of suggested commands.
const poetrySuggestionHeader = "Did you mean one of these perhaps?"

// poetryUnknownCommandMatch reports whether Poetry failed because the command
// does not exist and listed similar commands.
func poetryUnknownCommandMatch(command *types.Command) bool {
	if poetryUnknownCommand.FindStringSubmatch(command.Output) == nil {
		return false
	}
	return strings.Contains(command.Output, poetrySuggestionHeader)
}

// poetryUnknownCommandNewCommand replaces the unknown command with every
// command that Poetry suggests.
func poetryUnknownCommandNewCommand(command *types.Command) ([]string, error) {
	match := poetryUnknownCommand.FindStringSubmatch(command.Output)
	if match == nil {
		return nil, nil
	}
	return utils.ReplaceArguments(command.Script, match[1], poetrySuggestions(command.Output)), nil
}

// poetrySuggestions extracts the suggested command names from the indented
// list that follows the "Did you mean one of these perhaps?" line. Every entry
// has the form "    name: description".
func poetrySuggestions(commandOutput string) []string {
	suggestions := []string{}
	inSuggestions := false
	for _, line := range strings.Split(commandOutput, "\n") {
		if !inSuggestions {
			inSuggestions = strings.Contains(line, poetrySuggestionHeader)
			continue
		}
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if !strings.HasPrefix(line, "    ") {
			break
		}
		name, _, found := strings.Cut(trimmed, ":")
		if !found {
			continue
		}
		if name = strings.TrimSpace(name); name != "" {
			suggestions = append(suggestions, name)
		}
	}
	return suggestions
}

func init() {
	register(types.NewRule(
		"poetry_unknown_command",
		utils.ForApp([]string{"poetry"}, 0, poetryUnknownCommandMatch),
		poetryUnknownCommandNewCommand,
	))
}
