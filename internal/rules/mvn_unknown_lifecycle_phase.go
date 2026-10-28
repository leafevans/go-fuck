package rules

import (
	"regexp"
	"strings"

	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// mvnUnknownLifecyclePattern matches the unknown lifecycle phase reported by
// maven.
var mvnUnknownLifecyclePattern = regexp.MustCompile(`\[ERROR\] Unknown lifecycle phase "(.+)"`)

// mvnAvailableLifecyclePattern matches the list of available lifecycle phases.
var mvnAvailableLifecyclePattern = regexp.MustCompile(`Available lifecycle phases are: (.+) -> \[Help 1\]`)

// mvnUnknownLifecyclePhaseMatch reports whether maven reported an unknown
// lifecycle phase and listed the valid ones.
func mvnUnknownLifecyclePhaseMatch(command *types.Command) bool {
	return mvnUnknownLifecyclePattern.FindStringSubmatch(command.Output) != nil &&
		mvnAvailableLifecyclePattern.FindStringSubmatch(command.Output) != nil
}

// mvnUnknownLifecyclePhaseNewCommand replaces the unknown lifecycle phase with
// the closest valid one.
func mvnUnknownLifecyclePhaseNewCommand(command *types.Command) ([]string, error) {
	failed := mvnUnknownLifecyclePattern.FindStringSubmatch(command.Output)
	available := mvnAvailableLifecyclePattern.FindStringSubmatch(command.Output)
	if failed == nil || available == nil {
		return nil, nil
	}
	selected := utils.GetCloseMatches(failed[1], strings.Split(available[1], ", "), 0, 0.6)
	return utils.ReplaceCommand(command, failed[1], selected), nil
}

func init() {
	register(types.NewRule("mvn_unknown_lifecycle_phase",
		utils.ForApp([]string{"mvn"}, 0, mvnUnknownLifecyclePhaseMatch),
		mvnUnknownLifecyclePhaseNewCommand))
}
