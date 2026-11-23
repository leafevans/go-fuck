package rules

import (
	"github.com/leafevans/go-fuck/v3/internal/specific"
	"github.com/leafevans/go-fuck/v3/internal/types"
	"github.com/leafevans/go-fuck/v3/internal/utils"
)

// gitSwitchUnknownBranchReference matches the reference that git switch could
// not resolve.
var gitSwitchUnknownBranchReference = newRegexp(`fatal: invalid reference: (\S+)`)

// gitSwitchUnknownBranchMatch reports whether git switch failed because the
// branch does not exist.
func gitSwitchUnknownBranchMatch(command *types.Command) bool {
	return gitHasSubcommand(command.Script, "switch") &&
		gitSwitchUnknownBranchReference.MatchString(command.Output)
}

// gitSwitchUnknownBranchNewCommand suggests the closest existing branch or
// creating the missing one.
func gitSwitchUnknownBranchNewCommand(command *types.Command) ([]string, error) {
	match := gitSwitchUnknownBranchReference.FindStringSubmatch(command.Output)
	if match == nil {
		return nil, nil
	}
	missing := match[1]

	newCommands := []string{}
	if closest := utils.GetClosest(missing, gitCheckoutBranches(command.Script), 0.6, false); closest != "" {
		if corrected, ok := gitReplaceArgument(command.Script, missing, closest, false); ok {
			newCommands = append(newCommands, corrected)
		}
	}
	if corrected, ok := gitInsertAfterSubcommand(command.Script, "switch", "-c"); ok {
		newCommands = append(newCommands, corrected)
	}
	return newCommands, nil
}

func init() {
	register(types.NewRule("git_switch_unknown_branch",
		specific.GitSupportMatch(gitSwitchUnknownBranchMatch),
		specific.GitSupportNew(gitSwitchUnknownBranchNewCommand)))
}
