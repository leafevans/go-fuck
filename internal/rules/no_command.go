package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/internal/specific"
	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// Dependencies of the rule. They are variables so that tests do not depend on
// the installed executables or on the real shell history.
var (
	noCommandWhich             = utils.Which
	noCommandGetAllExecutables = utils.GetAllExecutables
	noCommandGetValidHistory   = utils.GetValidHistoryWithoutCurrent
)

// noCommandMatch reports whether the command executable does not exist but a
// similar executable does.
func noCommandMatch(command *types.Command) bool {
	parts := command.ScriptParts()
	if len(parts) == 0 || noCommandWhich(parts[0]) != "" {
		return false
	}

	if !strings.Contains(command.Output, "not found") &&
		!strings.Contains(command.Output, "is not recognized as") {
		return false
	}
	return len(utils.GetCloseMatches(parts[0], noCommandGetAllExecutables(), 0, 0.6)) > 0
}

// noCommandUsedExecutables returns the executables used in the history.
func noCommandUsedExecutables(command *types.Command) []string {
	executables := []string{}
	for _, script := range noCommandGetValidHistory(command) {
		executables = append(executables, strings.SplitN(script, " ", 2)[0])
	}
	return executables
}

// noCommandNewCommand replaces the missing executable with similar ones.
func noCommandNewCommand(command *types.Command) ([]string, error) {
	parts := command.ScriptParts()
	if len(parts) == 0 {
		return nil, nil
	}
	oldCommand := parts[0]

	// First try the executables that were already used in the history, then
	// all executables that are installed.
	newCommands := []string{}
	if alreadyUsed := utils.GetClosest(oldCommand, noCommandUsedExecutables(command), 0.6, false); alreadyUsed != "" {
		newCommands = append(newCommands, alreadyUsed)
	}
	for _, cmd := range utils.GetCloseMatches(oldCommand, noCommandGetAllExecutables(), 0, 0.6) {
		if !utils.Contains(newCommands, cmd) {
			newCommands = append(newCommands, cmd)
		}
	}

	results := make([]string, 0, len(newCommands))
	for _, cmd := range newCommands {
		results = append(results, strings.Replace(command.Script, oldCommand, cmd, 1))
	}
	return results, nil
}

func init() {
	rule := types.NewRule("no_command",
		specific.SudoSupportMatch(noCommandMatch),
		specific.SudoSupportNew(noCommandNewCommand))
	rule.Priority = 3000
	register(rule)
}
