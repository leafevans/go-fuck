package rules

import (
	"regexp"

	"github.com/leafevans/go-fuck/v2/internal/types"
	"github.com/leafevans/go-fuck/v2/internal/utils"
)

// cargoMissingBin matches the bin target that cargo could not find.
var cargoMissingBin = regexp.MustCompile("no bin target named `([^`]+)`")

// cargoBinTargetsMarker introduces the bin targets that cargo lists.
const cargoBinTargetsMarker = "available bin targets:"

// cargoMissingBinMatch reports whether cargo could not find a bin target.
func cargoMissingBinMatch(command *types.Command) bool {
	return cargoMissingBin.MatchString(command.Output) &&
		len(cargoAvailableTargets(command.Output, cargoBinTargetsMarker)) > 0
}

// cargoMissingBinNewCommand replaces the missing bin target with the closest
// one, or with the only one when a single target exists.
func cargoMissingBinNewCommand(command *types.Command) ([]string, error) {
	match := cargoMissingBin.FindStringSubmatch(command.Output)
	if match == nil {
		return nil, nil
	}
	missing := match[1]
	target := cargoClosestTarget(missing, cargoAvailableTargets(command.Output, cargoBinTargetsMarker))
	if target == "" {
		return nil, nil
	}
	return []string{utils.ReplaceArgument(command.Script, missing, target)}, nil
}

func init() {
	register(types.NewRule("cargo_missing_bin",
		utils.ForApp([]string{"cargo"}, 0, cargoMissingBinMatch),
		cargoMissingBinNewCommand))
}
