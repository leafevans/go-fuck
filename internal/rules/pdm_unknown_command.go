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
	if !strings.Contains(command.Output, "Command not found:") || !strings.Contains(command.Output, "Did you mean one of these commands?") {
		return "", nil
	}
	output := strings.ReplaceAll(pdmANSI.ReplaceAllString(command.Output, ""), "\r\n", "\n")
	match := pdmBrokenCommand.FindStringSubmatchIndex(output)
	if match == nil || !strings.Contains(output, "Did you mean one of these commands?") {
		return "", nil
	}
	lines := strings.Split(output[match[1]:], "\n")
	start := -1
	for index, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		if strings.TrimSpace(line) == "Did you mean one of these commands?" {
			start = index + 1
		}
		break
	}
	if start < 0 {
		return "", nil
	}
	var suggestions []string
	for _, line := range lines[start:] {
		if strings.Contains(line, "script commands") {
			break
		}
		if strings.TrimSpace(line) == "" {
			continue
		}
		parsed := pdmSuggestion.FindStringSubmatch(line)
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
