package rules

import (
	"regexp"
	"strings"

	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// awsCLIInvalidChoice matches the invalid aws choice and captures it.
//
// Deviation: Go regular expressions (RE2) do not support lookbehinds and
// lookaheads, so the Python pattern is replaced by an equivalent pattern that
// captures the mistake in a group.
var awsCLIInvalidChoice = regexp.MustCompile(`Invalid choice: '(.*)', maybe you meant:`)

// awsCLIOptions matches the choices suggested by the aws CLI.
var awsCLIOptions = regexp.MustCompile(`(?m)^\s*\*\s(.*)`)

// awsCLIMatch reports whether the aws CLI reported an invalid choice together
// with suggestions.
func awsCLIMatch(command *types.Command) bool {
	return strings.Contains(command.Output, "usage:") &&
		strings.Contains(command.Output, "maybe you meant:")
}

// awsCLINewCommand replaces the invalid aws choice with the suggested choices.
func awsCLINewCommand(command *types.Command) ([]string, error) {
	mistake := awsCLIInvalidChoice.FindStringSubmatch(command.Output)
	if mistake == nil {
		return nil, nil
	}
	options := awsCLIOptions.FindAllStringSubmatch(command.Output, -1)
	results := make([]string, 0, len(options))
	for _, option := range options {
		results = append(results, utils.ReplaceArgument(command.Script, mistake[1], option[1]))
	}
	return results, nil
}

func init() {
	register(types.NewRule("aws_cli",
		utils.ForApp([]string{"aws"}, 0, awsCLIMatch),
		awsCLINewCommand))
}
