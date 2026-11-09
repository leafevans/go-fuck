package rules

import (
	"regexp"
	"strings"

	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// The clap command line parser prints an error line and a "tip" line with the
// similar values it knows about. uv, cargo and rustup all use it, so the
// parsing is shared by their rules.
var (
	// clapUnrecognizedSubcommand matches the error line of an unknown
	// subcommand, for example `error: unrecognized subcommand 'instal'`.
	clapUnrecognizedSubcommand = regexp.MustCompile(`unrecognized subcommand '([^']*)'`)
	// clapUnexpectedArgument matches the error line of an unknown argument, for
	// example `error: unexpected argument '--versio' found`.
	clapUnexpectedArgument = regexp.MustCompile(`unexpected argument '([^']*)' found`)
	// clapQuotedSuggestion matches a quoted suggestion of a clap tip line, for
	// example `tip: a similar subcommand exists: 'install'`.
	clapQuotedSuggestion = regexp.MustCompile(`'([^']*)'`)
)

// clapSubcommandMatch reports whether clap reported an unknown subcommand
// together with at least one similar subcommand.
func clapSubcommandMatch(command *types.Command) bool {
	output := command.Output
	return strings.Contains(output, "unrecognized subcommand") &&
		strings.Contains(output, "tip:") &&
		strings.Contains(output, "similar subcommand")
}

// clapArgumentMatch reports whether clap reported an unknown argument together
// with at least one similar argument.
func clapArgumentMatch(command *types.Command) bool {
	output := command.Output
	return strings.Contains(output, "unexpected argument '") &&
		strings.Contains(output, "tip: a similar argument exists:")
}

// clapSubcommandSuggestions returns the misspelled subcommand and the similar
// subcommands from a clap error message.
func clapSubcommandSuggestions(commandOutput string) (string, []string) {
	return clapSuggestions(commandOutput, clapUnrecognizedSubcommand, "similar subcommand")
}

// clapArgumentSuggestions returns the misspelled argument and the similar
// arguments from a clap error message.
func clapArgumentSuggestions(commandOutput string) (string, []string) {
	return clapSuggestions(commandOutput, clapUnexpectedArgument, "similar argument")
}

// clapSuggestions extracts the broken value and the quoted suggestions of the
// tip lines that mention hint.
func clapSuggestions(commandOutput string, brokenPattern *regexp.Regexp, hint string) (string, []string) {
	match := brokenPattern.FindStringSubmatch(commandOutput)
	if match == nil {
		return "", nil
	}

	suggestions := []string{}
	for _, line := range strings.Split(commandOutput, newline) {
		if !strings.Contains(line, "tip:") || !strings.Contains(line, hint) {
			continue
		}
		for _, quoted := range clapQuotedSuggestion.FindAllStringSubmatch(line, -1) {
			suggestions = append(suggestions, quoted[1])
		}
	}
	return match[1], suggestions
}

// clapSubcommandNewCommand replaces the unknown subcommand with each suggestion
// of clap, ordered by similarity.
func clapSubcommandNewCommand(command *types.Command) ([]string, error) {
	broken, suggestions := clapSubcommandSuggestions(command.Output)
	if broken == "" {
		return nil, nil
	}
	return utils.ReplaceArgumentsSorted(command.Script, broken, suggestions), nil
}

// clapArgumentNewCommand replaces the unknown argument with each suggestion of
// clap, ordered by similarity.
func clapArgumentNewCommand(command *types.Command) ([]string, error) {
	broken, suggestions := clapArgumentSuggestions(command.Output)
	if broken == "" {
		return nil, nil
	}
	return utils.ReplaceArgumentsSorted(command.Script, broken, suggestions), nil
}
