package rules

import (
	"regexp"
	"strings"

	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// condaMistypeRegexp matches the quoted command names in the conda output.
var condaMistypeRegexp = regexp.MustCompile(`'conda ([^']*)'`)

// condaMistypeMatch reports whether conda suggested a corrected command.
func condaMistypeMatch(command *types.Command) bool {
	return strings.Contains(command.Output, "Did you mean 'conda")
}

// condaMistypeNewCommand replaces the misspelled command with the suggestion.
func condaMistypeNewCommand(command *types.Command) ([]string, error) {
	matches := condaMistypeRegexp.FindAllStringSubmatch(command.Output, -1)
	if len(matches) < 2 {
		return nil, nil
	}
	return utils.ReplaceCommand(command, matches[0][1], []string{matches[1][1]}), nil
}

func init() {
	register(types.NewRule("conda_mistype",
		utils.ForApp([]string{"conda"}, 0, condaMistypeMatch),
		condaMistypeNewCommand))
}
