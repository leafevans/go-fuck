package rules

import (
	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// historyGetValidHistory returns the history entries that may be suggested as
// corrections. It is a variable so that tests can inject a fixed history.
var historyGetValidHistory = utils.GetValidHistoryWithoutCurrent

// historyMatch reports whether a similar command is present in the history.
func historyMatch(command *types.Command) bool {
	similar := utils.GetCloseMatches(command.Script, historyGetValidHistory(command), 0, 0.6)
	return len(similar) > 0
}

// historyNewCommand suggests the history entry that is closest to the command.
func historyNewCommand(command *types.Command) ([]string, error) {
	script := utils.GetClosestDefault(command.Script, historyGetValidHistory(command))
	if script == "" {
		return nil, nil
	}
	return []string{script}, nil
}

func init() {
	rule := types.NewRule("history", historyMatch, historyNewCommand)
	rule.Priority = 9999
	register(rule)
}
