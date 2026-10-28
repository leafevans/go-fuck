package rules

import (
	"regexp"
	"strings"

	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

var (
	// cargoUnexpectedArgument matches the error line of cargo for an unknown
	// command line argument, for example: error: unexpected argument --relase found.
	cargoUnexpectedArgument = regexp.MustCompile("unexpected argument '([^']*)' found")
	// cargoSimilarArgument matches the help line that suggests an argument.
	cargoSimilarArgument = regexp.MustCompile("tip: a similar argument exists: '([^']*)'")
)

// cargoUnknownArgumentMatch matches the error that cargo prints for a
// misspelled command line argument when it can suggest a similar one.
func cargoUnknownArgumentMatch(command *types.Command) bool {
	output := command.Output
	return strings.Contains(output, "error: unexpected argument '") &&
		strings.Contains(output, "tip: a similar argument exists:")
}

// cargoUnknownArgumentNewCommand replaces the misspelled argument with each
// similar argument that cargo suggests.
func cargoUnknownArgumentNewCommand(command *types.Command) ([]string, error) {
	match := cargoUnexpectedArgument.FindStringSubmatch(command.Output)
	if match == nil {
		return nil, nil
	}
	broken := match[1]

	corrected := []string{}
	for _, suggestion := range cargoSimilarArgument.FindAllStringSubmatch(command.Output, -1) {
		corrected = append(corrected, utils.ReplaceArgument(command.Script, broken, suggestion[1]))
	}
	return corrected, nil
}

func init() {
	register(types.NewRule(
		"cargo_unknown_argument",
		utils.ForApp([]string{"cargo"}, 0, cargoUnknownArgumentMatch),
		cargoUnknownArgumentNewCommand,
	))
}
