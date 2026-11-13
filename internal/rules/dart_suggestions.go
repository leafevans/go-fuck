package rules

import (
	"regexp"
	"strings"

	"github.com/leafevans/go-fuck/v2/internal/types"
	"github.com/leafevans/go-fuck/v2/internal/utils"
)

// dartUnknownCommand matches the misspelled command reported by the args
// package CommandRunner, which the Dart CLI and the Flutter tool both use:
//
//	Could not find a command named "buidl".
//	Could not find a subcommand named "buidl" for "dart pub".
var dartUnknownCommand = regexp.MustCompile(`Could not find a (?:command|subcommand) named "([^"]*)"`)

// dartSimilarCommands extracts the misspelled command and the suggestions that
// the args package prints after a "Did you mean one of these?" line.
//
// The suggestions are two-space indented command names, for example:
//
//	Could not find a command named "buidl".
//
//	Did you mean one of these?
//	  build
//	  pub
//
//	Usage: dart <command|dart-file> [arguments]
func dartSimilarCommands(output string) (broken string, suggestions []string) {
	match := dartUnknownCommand.FindStringSubmatch(output)
	if match == nil {
		return "", nil
	}
	broken = match[1]

	inSuggestions := false
	for _, line := range strings.Split(output, newline) {
		trimmed := strings.TrimSpace(line)
		if trimmed == "Did you mean one of these?" {
			inSuggestions = true
			continue
		}
		if !inSuggestions {
			continue
		}
		if strings.HasPrefix(line, "  ") {
			suggestions = append(suggestions, trimmed)
			continue
		}
		if trimmed == "" {
			continue
		}
		break
	}
	return broken, suggestions
}

// dartUnknownCommandMatch reports whether a Dart or Flutter command is
// misspelled and the CLI printed at least one suggestion.
func dartUnknownCommandMatch(command *types.Command) bool {
	broken, suggestions := dartSimilarCommands(command.Output)
	return broken != "" && len(suggestions) > 0
}

// dartUnknownCommandNewCommand replaces the misspelled command with each
// suggestion printed by the args package.
func dartUnknownCommandNewCommand(command *types.Command) ([]string, error) {
	broken, suggestions := dartSimilarCommands(command.Output)
	if broken == "" {
		return nil, nil
	}
	return utils.ReplaceArguments(command.Script, broken, suggestions), nil
}
