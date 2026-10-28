package rules

import (
	"regexp"
	"strings"

	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// azCLIInvalidChoice matches the invalid az command and captures it.
//
// Deviation: Go regular expressions (RE2) do not support lookaheads. The
// leading ".*" reproduces the greedy search that the Python rule anchors with
// the "(?=az)" lookahead.
var azCLIInvalidChoice = regexp.MustCompile(`.*: '(.*)' is not in the '.*' command group.`)

// azCLIOptions matches the choices suggested by the az CLI.
var azCLIOptions = regexp.MustCompile(`(?m)^The most similar choice to '.*' is:\n\s*(.*)$`)

// azCLIMatch reports whether the az CLI reported an invalid command that is not
// in a command group.
func azCLIMatch(command *types.Command) bool {
	return strings.Contains(command.Output, "is not in the") &&
		strings.Contains(command.Output, "command group")
}

// azCLINewCommand replaces the invalid az command with the suggested choices.
func azCLINewCommand(command *types.Command) ([]string, error) {
	mistake := azCLIInvalidChoice.FindStringSubmatch(command.Output)
	if mistake == nil {
		return nil, nil
	}
	options := azCLIOptions.FindAllStringSubmatch(command.Output, -1)
	results := make([]string, 0, len(options))
	for _, option := range options {
		results = append(results, utils.ReplaceArgument(command.Script, mistake[1], option[1]))
	}
	return results, nil
}

func init() {
	register(types.NewRule("az_cli",
		utils.ForApp([]string{"az"}, 0, azCLIMatch),
		azCLINewCommand))
}
