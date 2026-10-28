package rules

import (
	"regexp"

	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// yarnCommandReplacedRunInstead matches the replacement command yarn suggests.
var yarnCommandReplacedRunInstead = regexp.MustCompile("Run \"(.*)\" instead")

// yarnCommandReplacedMatch reports whether yarn suggested running another
// command instead.
func yarnCommandReplacedMatch(command *types.Command) bool {
	return yarnCommandReplacedRunInstead.FindStringSubmatch(command.Output) != nil
}

// yarnCommandReplacedNewCommand runs the command yarn suggested.
func yarnCommandReplacedNewCommand(command *types.Command) ([]string, error) {
	match := yarnCommandReplacedRunInstead.FindStringSubmatch(command.Output)
	if match == nil {
		return nil, nil
	}
	return []string{match[1]}, nil
}

func init() {
	register(types.NewRule("yarn_command_replaced",
		utils.ForApp([]string{"yarn"}, 1, yarnCommandReplacedMatch),
		yarnCommandReplacedNewCommand))
}
