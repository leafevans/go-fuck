package rules

import (
	"regexp"
	"strings"

	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// cargoMissingTestTarget matches the test target that cargo could not find.
var cargoMissingTestTarget = regexp.MustCompile("no test target named `([^`]+)`")

// cargoTestTargetsMarker introduces the test targets that cargo lists.
const cargoTestTargetsMarker = "available test targets:"

// cargoMissingTestTargetMatch reports whether cargo could not find a test
// target.
func cargoMissingTestTargetMatch(command *types.Command) bool {
	parts := command.ScriptParts()
	if len(parts) < 2 || !strings.EqualFold(parts[1], "test") {
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
	return []string{utils.ReplaceArgument(command.Script, match[1], target)}, nil
}

func init() {
	register(types.NewRule("cargo_missing_test_target",
		utils.ForApp([]string{"cargo"}, 0, cargoMissingTestTargetMatch),
		cargoMissingTestTargetNewCommand))
}
