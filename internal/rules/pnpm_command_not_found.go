package rules

import (
	"regexp"
	"strings"

	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

var (
	// pnpmBrokenCommand matches the error line of pnpm for a command or
	// script that does not exist.
	pnpmBrokenCommand = regexp.MustCompile(`Command "([^"]*)" not found`)
	// pnpmSuggestion matches the "Did you mean" line of pnpm.
	pnpmSuggestion = regexp.MustCompile(`Did you mean "pnpm ([^"]*)"\?`)
)

// pnpmCommandNotFoundMatch matches the error that pnpm prints when a command
// or package script does not exist.
func pnpmCommandNotFoundMatch(command *types.Command) bool {
	output := command.Output
	return strings.Contains(output, "ERR_PNPM_RECURSIVE_EXEC_FIRST_FAIL") &&
		strings.Contains(output, `Command "`) &&
		strings.Contains(output, "not found")
}

// pnpmCommandNotFoundNewCommand replaces the unknown command with the
// suggestions that pnpm prints.
func pnpmCommandNotFoundNewCommand(command *types.Command) ([]string, error) {
	match := pnpmBrokenCommand.FindStringSubmatch(command.Output)
	if match == nil {
		return nil, nil
	}
	broken := match[1]

	matches := pnpmSuggestion.FindAllStringSubmatch(command.Output, -1)
	suggestions := make([]string, 0, len(matches))
	for _, match := range matches {
		suggestions = append(suggestions, match[1])
	}
	return utils.ReplaceArgumentsSorted(command.Script, broken, suggestions), nil
}

func init() {
	register(types.NewRule(
		"pnpm_command_not_found",
		utils.ForApp([]string{"pnpm"}, 0, pnpmCommandNotFoundMatch),
		pnpmCommandNotFoundNewCommand,
	))
}
