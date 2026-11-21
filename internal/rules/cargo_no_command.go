package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/v2/internal/types"
	"github.com/leafevans/go-fuck/v2/internal/utils"
)

// cargoNoCommandSuggestionRegexps match the subcommand suggested by cargo.
//
// Old cargo releases printed "no such subcommand" followed by "Did you mean"
// and the subcommand; current releases print "no such command" and a help line
// with the similar name. Both forms have to be understood, otherwise the rule
// silently stops working after a cargo upgrade.
var cargoNoCommandSuggestionRegexps = []*lazyRegexp{
	newRegexp("Did you mean `([^`]*)`"),
	newRegexp("a command with a similar name exists: `([^`]*)`"),
	newRegexp("a similar subcommand exists: [`']([^`']*)[`']"),
}

var cargoUnrecognizedSubcommand = newRegexp("unrecognized subcommand [`']([^`']+)[`']")

// cargoNoCommandSuggestion returns the subcommand suggested by cargo, if the
// output carries one.
func cargoNoCommandSuggestion(output string) string {
	for _, pattern := range cargoNoCommandSuggestionRegexps {
		if match := pattern.FindStringSubmatch(output); match != nil {
			return match[1]
		}
	}
	return ""
}

// cargoNoCommandMatch reports whether cargo did not recognize the subcommand.
func cargoNoCommandMatch(command *types.Command) bool {
	output := strings.ToLower(command.Output)
	if !strings.Contains(output, "no such subcommand") &&
		!strings.Contains(output, "no such command") &&
		!strings.Contains(output, "unrecognized subcommand") {
		return false
	}
	return cargoNoCommandSuggestion(command.Output) != ""
}

// cargoNoCommandNewCommand replaces the misspelled subcommand.
func cargoNoCommandNewCommand(command *types.Command) ([]string, error) {
	words := simpleCommandWords(command.Script)
	index := cargoSubcommandIndex(words)
	if index < 0 {
		return nil, nil
	}
	expected := words[index].text
	for _, pattern := range []*lazyRegexp{cargoNoSuchCommand, cargoUnrecognizedSubcommand} {
		if match := pattern.FindStringSubmatch(command.Output); match != nil {
			expected = match[1]
			break
		}
	}
	return replaceSubcommand(command.Script, words, index, expected, cargoNoCommandSuggestion(command.Output)), nil
}

func init() {
	register(types.NewRule("cargo_no_command",
		utils.ForApp([]string{"cargo"}, 1, cargoNoCommandMatch),
		cargoNoCommandNewCommand))
}
