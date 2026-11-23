package rules

import (
	"github.com/leafevans/go-fuck/v3/internal/types"
	"github.com/leafevans/go-fuck/v3/internal/utils"
)

// cargoMissingBenchTarget matches the bench target that cargo could not find.
var cargoMissingBenchTarget = newRegexp("no bench target named `([^`]+)`")

// cargoBenchTargetsMarker introduces the bench targets that cargo lists.
const cargoBenchTargetsMarker = "available bench targets:"

// cargoMissingBenchTargetMatch reports whether cargo could not find a bench
// target.
func cargoMissingBenchTargetMatch(command *types.Command) bool {
	words := simpleCommandWords(command.Script)
	if index := cargoSubcommandIndex(words); index < 0 || words[index].text != "bench" {
		return false
	}
	return cargoMissingBenchTarget.MatchString(command.Output) &&
		len(cargoAvailableTargets(command.Output, cargoBenchTargetsMarker)) > 0
}

// cargoMissingBenchTargetNewCommand replaces the missing bench target with the
// closest one.
func cargoMissingBenchTargetNewCommand(command *types.Command) ([]string, error) {
	match := cargoMissingBenchTarget.FindStringSubmatch(command.Output)
	if match == nil {
		return nil, nil
	}
	target := cargoClosestTarget(match[1],
		cargoAvailableTargets(command.Output, cargoBenchTargetsMarker))
	if target == "" {
		return nil, nil
	}
	return cargoReplaceTarget(command.Script, "--bench", match[1], target), nil
}

func init() {
	register(types.NewRule("cargo_missing_bench_target",
		utils.ForApp([]string{"cargo"}, 0, cargoMissingBenchTargetMatch),
		cargoMissingBenchTargetNewCommand))
}
