package rules

import (
	"regexp"
	"strings"

	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

var (
	// rustupUnrecognizedSubcommand matches the error line of rustup, for example
	// `error: unrecognized subcommand 'ad'`.
	rustupUnrecognizedSubcommand = regexp.MustCompile(`unrecognized subcommand '([^']*)'`)
	// rustupQuotedSuggestion matches the quoted suggestions of the rustup "tip"
	// line, for example `tip: a similar subcommand exists: 'add'`.
	rustupQuotedSuggestion = regexp.MustCompile(`'([^']*)'`)
)

// rustupUnknownSubcommandMatch reports whether rustup reported an unknown
// subcommand together with at least one similar subcommand.
func rustupUnknownSubcommandMatch(command *types.Command) bool {
	output := command.Output
	return strings.Contains(output, "error: unrecognized subcommand") &&
		strings.Contains(output, "tip:") &&
		strings.Contains(output, "similar subcommand")
}

// rustupUnknownSubcommandNewCommand replaces the misspelled subcommand with
// each suggestion of rustup.
func rustupUnknownSubcommandNewCommand(command *types.Command) ([]string, error) {
	match := rustupUnrecognizedSubcommand.FindStringSubmatch(command.Output)
	if match == nil {
		return nil, nil
	}
	broken := match[1]

	suggestions := []string{}
	for _, line := range strings.Split(command.Output, newline) {
		if !strings.Contains(line, "tip:") || !strings.Contains(line, "similar subcommand") {
			continue
		}
		for _, quoted := range rustupQuotedSuggestion.FindAllStringSubmatch(line, -1) {
			suggestions = append(suggestions, quoted[1])
		}
	}
	return utils.ReplaceArguments(command.Script, broken, utils.SortBySimilarity(broken, suggestions)), nil
}

func init() {
	register(types.NewRule(
		"rustup_unknown_subcommand",
		utils.ForApp([]string{"rustup"}, 0, rustupUnknownSubcommandMatch),
		rustupUnknownSubcommandNewCommand,
	))
}
