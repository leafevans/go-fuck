package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/v4/internal/types"
	"github.com/leafevans/go-fuck/v4/internal/utils"
)

var (
	// cargoNoSuchCommand matches the error line of cargo for an unknown
	// subcommand, for example: error: no such command: `instll`.
	cargoNoSuchCommand = newRegexp("no such command: \x60([^\x60]*)\x60")
	// cargoSimilarCommand matches the help line that suggests a command.
	cargoSimilarCommand = newRegexp("a command with a similar name exists: \x60([^\x60]*)\x60")
)

// cargoUnknownCommandMatch matches the error that cargo prints for a
// misspelled subcommand when it can suggest a similar command.
func cargoUnknownCommandMatch(command *types.Command) bool {
	output := command.Output
	return strings.Contains(output, "error: no such command: ") &&
		strings.Contains(output, "a command with a similar name exists")
}

// cargoUnknownCommandNewCommand replaces the misspelled subcommand with each
// similar command that cargo suggests.
func cargoUnknownCommandNewCommand(command *types.Command) ([]string, error) {
	match := cargoNoSuchCommand.FindStringSubmatch(command.Output)
	if match == nil {
		return nil, nil
	}
	broken := match[1]

	matches := cargoSimilarCommand.FindAllStringSubmatch(command.Output, -1)
	suggestions := make([]string, 0, len(matches))
	for _, match := range matches {
		suggestions = append(suggestions, match[1])
	}
	words := simpleCommandWords(command.Script)
	return replaceSubcommand(command.Script, words, cargoSubcommandIndex(words), broken, suggestions...), nil
}

func init() {
	register(types.NewRule(
		"cargo_unknown_command",
		utils.ForApp([]string{"cargo"}, 0, cargoUnknownCommandMatch),
		cargoUnknownCommandNewCommand,
	))
}
