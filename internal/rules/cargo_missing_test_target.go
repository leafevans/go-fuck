package rules

import (
	"github.com/leafevans/go-fuck/v3/internal/types"
	"github.com/leafevans/go-fuck/v3/internal/utils"
)

// cargoMissingTestTarget matches the test target that cargo could not find.
var cargoMissingTestTarget = newRegexp("no test target named `([^`]+)`")

// cargoTestTargetsMarker introduces the test targets that cargo lists.
const cargoTestTargetsMarker = "available test targets:"

// cargoMissingTestTargetMatch reports whether cargo could not find a test
// target.
func cargoMissingTestTargetMatch(command *types.Command) bool {
	words := simpleCommandWords(command.Script)
	if index := cargoSubcommandIndex(words); index < 0 || words[index].text != "test" {
		return false
	}
	return cargoMissingTestTarget.MatchString(command.Output) &&
		len(cargoAvailableTargets(command.Output, cargoTestTargetsMarker)) > 0
}

// cargoMissingTestTargetNewCommand replaces the missing test target with the
// closest one.
func cargoMissingTestTargetNewCommand(command *types.Command) ([]string, error) {
	match := cargoMissingTestTarget.FindStringSubmatch(command.Output)
	if match == nil {
		return nil, nil
	}
	target := cargoClosestTarget(match[1],
		cargoAvailableTargets(command.Output, cargoTestTargetsMarker))
	if target == "" {
		return nil, nil
	}
	return cargoReplaceTarget(command.Script, "--test", match[1], target), nil
}

func init() {
	register(types.NewRule("cargo_missing_test_target",
		utils.ForApp([]string{"cargo"}, 0, cargoMissingTestTargetMatch),
		cargoMissingTestTargetNewCommand))
}
