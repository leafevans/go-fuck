package rules

import (
	"regexp"

	"github.com/leafevans/go-fuck/internal/specific"
	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// gitSwitchUnknownBranchReference matches the reference that git switch could
// not resolve.
var gitSwitchUnknownBranchReference = regexp.MustCompile(`fatal: invalid reference: (\S+)`)

// gitSwitchUnknownBranchMatch reports whether git switch failed because the
// branch does not exist.
func gitSwitchUnknownBranchMatch(command *types.Command) bool {
	parts := command.ScriptParts()
	if len(parts) < 2 || parts[1] != "switch" {
		return false
	}
	return gitSwitchUnknownBranchReference.MatchString(command.Output)
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
	if closest := utils.GetClosest(missing, gitCheckoutBranches(), 0.6, false); closest != "" {
		newCommands = append(newCommands, utils.ReplaceArgument(command.Script, missing, closest))
	}
	newCommands = append(newCommands,
		utils.ReplaceArgument(command.Script, "switch", "switch -c"))
	return newCommands, nil
}

func init() {
	register(types.NewRule("git_switch_unknown_branch",
		specific.GitSupportMatch(gitSwitchUnknownBranchMatch),
		specific.GitSupportNew(gitSwitchUnknownBranchNewCommand)))
}
