package rules

import (
	"regexp"
	"strings"
)

// cobraUnknownCommand matches the error line printed by command line tools that
// use the Cobra framework, for example `unknown command "prr" for "gh"`.
var cobraUnknownCommand = regexp.MustCompile(`unknown command "([^"]*)" for "([^"]*)"`)

// cobraSuggestions extracts the broken command and the "Did you mean this?"
// suggestions from a Cobra error message.
//
// Cobra prints the suggestions as tab indented lines directly after the
// separator, followed by the usage text, for example:
//
//	unknown command "prr" for "gh"
//
//	Did you mean this?
//		org
//		pr
//	Usage:  gh <command> <subcommand> [flags]
func cobraSuggestions(commandOutput string) (broken string, suggestions []string) {
	match := cobraUnknownCommand.FindStringSubmatch(commandOutput)
	if match == nil {
		return "", nil
	}
	broken = match[1]

	inSuggestions := false
	for _, line := range strings.Split(commandOutput, newline) {
		trimmed := strings.TrimSpace(line)
		if trimmed == "Did you mean this?" {
			inSuggestions = true
			continue
		}
		if !inSuggestions {
			continue
		}
		if strings.HasPrefix(line, "	") {
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
