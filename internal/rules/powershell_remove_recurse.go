package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/v2/internal/types"
)

// The prompt and the non-interactive error that PowerShell prints when
// Remove-Item has to remove a directory with children without -Recurse.
const (
	powerShellRemoveRecursePrompt = "has children and the Recurse parameter was not specified"
	powerShellPromptUnavailable   = "NonInteractive mode. Read and Prompt functionality is not available"
)

// powerShellRemoveCommands are the PowerShell commands that remove items.
var powerShellRemoveCommands = []string{"rm", "del", "erase", "rd", "rmdir", "remove-item", "ri"}

// powerShellRemoveRecurseMatch reports whether a removal command needs the
// Recurse parameter.
func powerShellRemoveRecurseMatch(command *types.Command) bool {
	if !powerShellCommandIs(command, powerShellRemoveCommands...) {
		return false
	}
	output := command.Output
	return strings.Contains(output, powerShellRemoveRecursePrompt) ||
		strings.Contains(output, powerShellPromptUnavailable)
}

// powerShellRemoveRecurseNewCommand adds the Recurse parameter.
func powerShellRemoveRecurseNewCommand(command *types.Command) ([]string, error) {
	if len(command.ScriptParts()) == 0 {
		return nil, nil
	}
	return []string{powerShellInsertArgument(command.Script, "-Recurse")}, nil
}

func init() {
	register(types.NewRule("powershell_remove_recurse",
		powerShellRemoveRecurseMatch,
		powerShellRemoveRecurseNewCommand))
}
