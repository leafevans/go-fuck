package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/internal/specific"
	"github.com/leafevans/go-fuck/internal/types"
)

// pythonCommandMatch reports whether a python script was executed directly but
// could not be run.
func pythonCommandMatch(command *types.Command) bool {
	parts := command.ScriptParts()
	return len(parts) > 0 &&
		strings.HasSuffix(parts[0], ".py") &&
		(strings.Contains(command.Output, "Permission denied") ||
			strings.Contains(command.Output, "command not found"))
}

// pythonCommandNewCommand runs the script with python.
func pythonCommandNewCommand(command *types.Command) ([]string, error) {
	return []string{"python " + command.Script}, nil
}

func init() {
	register(types.NewRule("python_command",
		specific.SudoSupportMatch(pythonCommandMatch),
		specific.SudoSupportNew(pythonCommandNewCommand)))
}
