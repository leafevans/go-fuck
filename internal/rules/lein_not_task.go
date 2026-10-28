package rules

import (
	"regexp"
	"strings"

	"github.com/leafevans/go-fuck/internal/specific"
	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// leinNotTaskPattern extracts the unknown lein task from the output.
var leinNotTaskPattern = regexp.MustCompile("'([^']*)' is not a task")

// leinNotTaskMatch reports whether lein reported an unknown task together with
// suggestions.
func leinNotTaskMatch(command *types.Command) bool {
	return strings.HasPrefix(command.Script, "lein") &&
		strings.Contains(command.Output, "is not a task. See 'lein help'") &&
		strings.Contains(command.Output, "Did you mean this?")
}

// leinNotTaskNewCommand replaces the unknown task with the suggested ones.
func leinNotTaskNewCommand(command *types.Command) ([]string, error) {
	broken := leinNotTaskPattern.FindStringSubmatch(command.Output)
	if broken == nil {
		return nil, nil
	}
	matched := utils.GetAllMatchedCommands(command.Output, "Did you mean this?")
	return utils.ReplaceCommand(command, broken[1], matched), nil
}

func init() {
	register(types.NewRule("lein_not_task",
		specific.SudoSupportMatch(utils.ForApp([]string{"lein"}, 0, leinNotTaskMatch)),
		specific.SudoSupportNew(leinNotTaskNewCommand)))
}
