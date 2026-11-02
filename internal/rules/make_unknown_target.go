package rules

import (
	"regexp"

	"github.com/leafevans/go-fuck/internal/specific"
	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// makeUnknownTarget matches the GNU make error for a target that is not in the
// Makefile, for example: make: *** No rule to make target 'buidl'.  Stop.
var makeUnknownTarget = regexp.MustCompile("No rule to make target [\x60']([^\x60']*)[\x60']")

// makeUnknownTargetMatch reports whether make failed because the target does
// not exist and the target is part of the command line. The "needed by"
// variant of the error names a dependency that cannot be corrected by
// replacing an argument, so it is ignored.
func makeUnknownTargetMatch(command *types.Command) bool {
	match := makeUnknownTarget.FindStringSubmatch(command.Output)
	if match == nil || len(specific.MakeTargets()) == 0 {
		return false
	}
	return utils.Contains(command.ScriptParts(), match[1])
}

// makeUnknownTargetNewCommand suggests the Makefile targets that are close to
// the misspelled one.
func makeUnknownTargetNewCommand(command *types.Command) ([]string, error) {
	match := makeUnknownTarget.FindStringSubmatch(command.Output)
	if match == nil {
		return nil, nil
	}
	return utils.ReplaceCommand(command, match[1], specific.MakeTargets()), nil
}

func init() {
	register(types.NewRule(
		"make_unknown_target",
		utils.ForApp([]string{"make", "gmake"}, 0, makeUnknownTargetMatch),
		makeUnknownTargetNewCommand,
	))
}
