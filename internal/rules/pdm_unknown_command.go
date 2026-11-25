package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/v4/internal/types"
	"github.com/leafevans/go-fuck/v4/internal/utils"
)

// pdmUnknownCommand formats the unknown-command message with Rich markup, so
// the captured output may keep ANSI color sequences; strip them before parsing.
var pdmANSI = newRegexp("\\x1b\\[[0-9;]*[A-Za-z]")

// pdmBrokenCommand matches the command that pdm did not recognize.
var pdmBrokenCommand = newRegexp("Command not found: ([A-Za-z0-9_.+-]+)")

// pdmSuggestion matches one suggestion line of pdm's unknown-command message.
var pdmSuggestion = newRegexp("^  - ([A-Za-z0-9_.+-]+)$")

// pdmUnknownCommandParsed returns the unknown command and the subcommands that
// pdm suggested. Only the subcommand section is used; project script names are
// invoked with "pdm run" and are not valid replacements for a subcommand.
func pdmUnknownCommandParsed(command *types.Command) (string, []string) {
	output := pdmANSI.ReplaceAllString(command.Output, "")
	match := pdmBrokenCommand.FindStringSubmatch(output)
	if match == nil || !strings.Contains(output, "Did you mean one of these commands?") {
		return "", nil
	}
	lines := strings.Split(output, "\n")
	start := -1
	for index, line := range lines {
		if strings.TrimSpace(line) == "Did you mean one of these commands?" {
			start = index + 1
			break
		}
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
	return match[1], suggestions
}

func pdmUnknownCommandMatch(command *types.Command) bool {
	broken, suggestions := pdmUnknownCommandParsed(command)
	return broken != "" && len(suggestions) > 0
}

func pdmUnknownCommandNewCommand(command *types.Command) ([]string, error) {
	broken, suggestions := pdmUnknownCommandParsed(command)
	if broken == "" {
		return nil, nil
	}
	return utils.ReplaceArgumentsSorted(command.Script, broken, suggestions), nil
}

func init() {
	register(types.NewRule(
		"pdm_unknown_command",
		utils.ForApp([]string{"pdm"}, 0, pdmUnknownCommandMatch),
		pdmUnknownCommandNewCommand,
	))
}
