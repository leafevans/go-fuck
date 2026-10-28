package rules

import (
	"regexp"
	"strings"

	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

var (
	// uvUnrecognizedSubcommand matches the error line of uv, for example
	// `error: unrecognized subcommand 'instal'`.
	uvUnrecognizedSubcommand = regexp.MustCompile(`unrecognized subcommand '([^']*)'`)
	// uvQuotedSuggestion matches the quoted suggestions of the uv "tip" line.
	uvQuotedSuggestion = regexp.MustCompile(`'([^']*)'`)
)

// uvUnknownSubcommandMatch matches the typo error that uv prints when a
// subcommand is misspelled.
func uvUnknownSubcommandMatch(command *types.Command) bool {
	output := command.Output
	return strings.Contains(output, "unrecognized subcommand") &&
		strings.Contains(output, "tip:") &&
		strings.Contains(output, "similar subcommand")
}

// uvUnknownSubcommandNewCommand replaces the misspelled subcommand with each
// suggestion of uv.
func uvUnknownSubcommandNewCommand(command *types.Command) ([]string, error) {
	match := uvUnrecognizedSubcommand.FindStringSubmatch(command.Output)
	if match == nil {
		return nil, nil
	}
	broken := match[1]

	corrected := []string{}
	for _, line := range strings.Split(command.Output, newline) {
		if !strings.Contains(line, "tip:") || !strings.Contains(line, "similar subcommand") {
			continue
		}
		for _, quoted := range uvQuotedSuggestion.FindAllStringSubmatch(line, -1) {
			corrected = append(corrected, utils.ReplaceArgument(command.Script, broken, quoted[1]))
		}
	}
	return corrected, nil
}

func init() {
	register(types.NewRule(
		"uv_unknown_subcommand",
		utils.ForApp([]string{"uv"}, 0, uvUnknownSubcommandMatch),
		uvUnknownSubcommandNewCommand,
	))
}
