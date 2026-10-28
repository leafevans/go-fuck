package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// pythonExecuteSuffix is the suffix that python expects for script files.
const pythonExecuteSuffix = ".py"

// pythonExecuteMatch reports whether python was called without a .py suffix.
func pythonExecuteMatch(command *types.Command) bool {
	return !strings.HasSuffix(command.Script, pythonExecuteSuffix)
}

// pythonExecuteNewCommand appends the missing suffix.
func pythonExecuteNewCommand(command *types.Command) ([]string, error) {
	return []string{command.Script + pythonExecuteSuffix}, nil
}

func init() {
	register(types.NewRule("python_execute",
		utils.ForApp([]string{"python"}, 0, pythonExecuteMatch),
		pythonExecuteNewCommand))
}
