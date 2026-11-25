package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/v4/internal/types"
)

// pdmUnknownCommand formats the unknown-command message with Rich markup, so
// the captured output may keep ANSI color sequences; strip them before parsing.
var pdmANSI = newRegexp("\\x1b\\[[0-9;]*[A-Za-z]")

// pdmBrokenCommand matches the command that pdm did not recognize.
var pdmBrokenCommand = newRegexp("(?m)^pdm: error: Command not found: ([A-Za-z0-9][A-Za-z0-9_.+-]*)\\r?$")

// pdmSuggestion matches one suggestion line of pdm's unknown-command message.
var pdmSuggestion = newRegexp("^[ \t]+- ([A-Za-z0-9][A-Za-z0-9_.+-]*)$")

// pdmUnknownCommandParsed returns the unknown command and the subcommands that
// pdm suggested. Only the subcommand section is used; project script names are
// invoked with "pdm run" and are not valid replacements for a subcommand.
func pdmUnknownCommandParsed(command *types.Command) (string, []string) {
	const header = "Did you mean one of these commands?"
	output := command.Output
	if strings.ContainsRune(output, '\x1b') {
		output = pdmANSI.ReplaceAllString(output, "")
	}
	if !strings.Contains(output, "Command not found:") || !strings.Contains(output, header) {
		return "", nil
	}
	match := pdmBrokenCommand.FindStringSubmatchIndex(output)
	if match == nil {
		return "", nil
	}
	var suggestions []string
	inSuggestions := false
	for line := range strings.SplitSeq(output[match[1]:], "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if !inSuggestions {
			if trimmed != header {
				return "", nil
			}
			inSuggestions = true
			continue
		}
		parsed := pdmSuggestion.FindStringSubmatch(strings.TrimSuffix(line, "\r"))
		if parsed == nil {
			break
		}
		suggestions = append(suggestions, parsed[1])
	}
	return output[match[2]:match[3]], suggestions
}

func pdmCorrections(command *types.Command) []string {
	broken, suggestions := pdmUnknownCommandParsed(command)
	if len(suggestions) == 0 {
		return nil
	}
	return cliSubcommandCorrections(command, []string{"pdm", "pdm.exe"}, nil, broken, suggestions, func(name string) commandOption {
		switch name {
		case "-p", "--project":
			return valueOption
		case "--no-cache", "-v", "-vv", "--verbose", "--non-interactive":
			return flagOption
		default:
			return unknownOption
		}
	})
}

func init() {
	register(types.NewRule("pdm_unknown_command", func(command *types.Command) bool { return len(pdmCorrections(command)) > 0 },
		func(command *types.Command) ([]string, error) { return pdmCorrections(command), nil }))
}
