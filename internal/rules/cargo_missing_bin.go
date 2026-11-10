package rules

import (
	"regexp"
	"strings"

	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// cargoMissingBin matches the bin target that cargo could not find.
var cargoMissingBin = regexp.MustCompile("no bin target named `([^`]+)`")

// cargoBinTargets returns the bin targets listed by the cargo help output.
func cargoBinTargets(output string) []string {
	const marker = "available bin targets:"
	index := strings.Index(output, marker)
	if index < 0 {
		return nil
	}
	var targets []string
	for _, line := range strings.Split(output[index+len(marker):], "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if !strings.HasPrefix(line, "    ") {
			break
		}
		targets = append(targets, trimmed)
	}
	return targets
}

// cargoMissingBinMatch reports whether cargo could not find a bin target.
func cargoMissingBinMatch(command *types.Command) bool {
	return cargoMissingBin.MatchString(command.Output) && len(cargoBinTargets(command.Output)) > 0
}

// cargoMissingBinNewCommand replaces the missing bin target with the closest
// one, or with the only one when a single target exists.
func cargoMissingBinNewCommand(command *types.Command) ([]string, error) {
	match := cargoMissingBin.FindStringSubmatch(command.Output)
	if match == nil {
		return nil, nil
	}
	missing := match[1]
	targets := cargoBinTargets(command.Output)
	target := utils.GetClosest(missing, targets, 0.6, false)
	if target == "" && len(targets) == 1 {
		target = targets[0]
	}
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
