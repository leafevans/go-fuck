package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/v4/internal/shells"
	"github.com/leafevans/go-fuck/v4/internal/types"
)

// pythonModuleErrorMissingModule matches the missing module reported by python.
var pythonModuleErrorMissingModule = newRegexp(`ModuleNotFoundError: No module named '([^']+)'`)

// pythonModuleErrorMatch reports whether python failed to import a module.
func pythonModuleErrorMatch(command *types.Command) bool {
	return strings.Contains(command.Output, "ModuleNotFoundError: No module named '")
}

// pythonModuleErrorNewCommand installs the missing module and retries.
func pythonModuleErrorNewCommand(command *types.Command) ([]string, error) {
	match := pythonModuleErrorMissingModule.FindStringSubmatch(command.Output)
	if match == nil {
		return nil, nil
	}
	return []string{shells.Current().And("pip install "+match[1], command.Script)}, nil
}

func init() {
	register(types.NewRule("python_module_error",
		pythonModuleErrorMatch,
		pythonModuleErrorNewCommand))
}
