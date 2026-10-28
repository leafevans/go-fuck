package rules

import (
	"regexp"
	"strings"

	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

var (
	// uvUnexpectedArgument matches the error line of uv for an unknown command
	// line argument, for example: error: unexpected argument --versio found.
	uvUnexpectedArgument = regexp.MustCompile(`unexpected argument '([^']*)' found`)
	// uvSimilarArgument matches the clap tip line that suggests an argument.
	uvSimilarArgument = regexp.MustCompile(`tip: a similar argument exists: '([^']*)'`)
)

// uvUnknownArgumentMatch matches the error that uv prints for a misspelled
// command line argument when it can suggest a similar one.
func uvUnknownArgumentMatch(command *types.Command) bool {
	output := command.Output
	return strings.Contains(output, "error: unexpected argument '") &&
		strings.Contains(output, "tip: a similar argument exists:")
}

// uvUnknownArgumentNewCommand replaces the misspelled argument with each
// similar argument that uv suggests.
func uvUnknownArgumentNewCommand(command *types.Command) ([]string, error) {
	match := uvUnexpectedArgument.FindStringSubmatch(command.Output)
	if match == nil {
		return nil, nil
	}
	broken := match[1]

	corrected := []string{}
	for _, suggestion := range uvSimilarArgument.FindAllStringSubmatch(command.Output, -1) {
		corrected = append(corrected, utils.ReplaceArgument(command.Script, broken, suggestion[1]))
	}
	return corrected, nil
}

func init() {
	register(types.NewRule(
		"uv_unknown_argument",
		utils.ForApp([]string{"uv"}, 0, uvUnknownArgumentMatch),
		uvUnknownArgumentNewCommand,
	))
}
