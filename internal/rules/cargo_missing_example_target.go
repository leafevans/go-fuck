package rules

import (
	"regexp"

	"github.com/leafevans/go-fuck/v2/internal/types"
	"github.com/leafevans/go-fuck/v2/internal/utils"
)

// cargoMissingExampleTarget matches the example target that cargo could not
// find.
var cargoMissingExampleTarget = regexp.MustCompile("no example target named `([^`]+)`")

// cargoExampleTargetsMarker introduces the example targets that cargo lists.
const cargoExampleTargetsMarker = "available example targets:"

// cargoMissingExampleTargetMatch reports whether cargo could not find an
// example target.
func cargoMissingExampleTargetMatch(command *types.Command) bool {
	return cargoMissingExampleTarget.MatchString(command.Output) &&
		len(cargoAvailableTargets(command.Output, cargoExampleTargetsMarker)) > 0
}

// cargoMissingExampleTargetNewCommand replaces the missing example target with
// the closest one.
func cargoMissingExampleTargetNewCommand(command *types.Command) ([]string, error) {
	match := cargoMissingExampleTarget.FindStringSubmatch(command.Output)
	if match == nil {
		return nil, nil
	}
	target := cargoClosestTarget(match[1],
		cargoAvailableTargets(command.Output, cargoExampleTargetsMarker))
	if target == "" {
		return nil, nil
	}
	return []string{utils.ReplaceArgument(command.Script, match[1], target)}, nil
}

func init() {
	register(types.NewRule("cargo_missing_example_target",
		utils.ForApp([]string{"cargo"}, 0, cargoMissingExampleTargetMatch),
		cargoMissingExampleTargetNewCommand))
}
