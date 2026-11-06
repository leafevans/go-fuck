package rules

import (
	"regexp"
	"strings"

	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// dotnetUnknownArgumentHeader starts the suggestion block that the .NET CLI
// System.CommandLine parser prints for a misspelled nested command, for
// example:
//
//	'instal' was not matched. Did you mean one of the following?
var dotnetUnknownArgumentHeader = "was not matched. Did you mean one of the following?"

// dotnetUnknownArgumentBroken matches the misspelled argument on the header
// line, for example: 'instal' was not matched. The "Unrecognized command or
// argument" line below it is localized, so it must not be used to extract the
// argument.
var dotnetUnknownArgumentBroken = regexp.MustCompile("'([^']*)' was not matched")

// dotnetUnknownArgumentSuggestions returns the suggestions printed after the
// "was not matched" header line.
func dotnetUnknownArgumentSuggestions(output string) []string {
	suggestions := []string{}
	inSuggestions := false
	for _, line := range strings.Split(output, newline) {
		trimmed := strings.TrimSpace(line)
		if strings.Contains(line, dotnetUnknownArgumentHeader) {
			inSuggestions = true
			continue
		}
		if !inSuggestions {
			continue
		}
		if trimmed == "" {
			break
		}
		suggestions = append(suggestions, trimmed)
	}
	return suggestions
}

// dotnetUnknownArgumentMatch reports whether a nested .NET CLI command is
// misspelled and the CLI printed at least one suggestion.
func dotnetUnknownArgumentMatch(command *types.Command) bool {
	if !strings.Contains(command.Output, dotnetUnknownArgumentHeader) {
		return false
	}
	if dotnetUnknownArgumentBroken.FindStringSubmatch(command.Output) == nil {
		return false
	}
	return len(dotnetUnknownArgumentSuggestions(command.Output)) > 0
}

// dotnetUnknownArgumentNewCommand replaces the misspelled nested command with
// each suggestion printed by the .NET CLI.
func dotnetUnknownArgumentNewCommand(command *types.Command) ([]string, error) {
	match := dotnetUnknownArgumentBroken.FindStringSubmatch(command.Output)
	if match == nil {
		return nil, nil
	}
	return utils.ReplaceArguments(command.Script, match[1], dotnetUnknownArgumentSuggestions(command.Output)), nil
}

func init() {
	register(types.NewRule(
		"dotnet_unknown_argument",
		utils.ForApp([]string{"dotnet"}, 0, dotnetUnknownArgumentMatch),
		dotnetUnknownArgumentNewCommand,
	))
}
