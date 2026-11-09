package rules

import (
	"regexp"

	"github.com/leafevans/go-fuck/internal/types"
)

// powerShellExistingItem matches the existing item that New-Item reports.
var powerShellExistingItem = regexp.MustCompile(`An item with the specified name (.+) already exists\.`)

// powerShellMkdirExistingMatch reports whether mkdir failed because the
// directory already exists.
func powerShellMkdirExistingMatch(command *types.Command) bool {
	if !powerShellCommandIs(command, "mkdir", "md") {
		return false
	}
	match := powerShellExistingItem.FindStringSubmatch(command.Output)
	if match == nil {
		return false
	}
	return powerShellItemIsDirectory(match[1])
}

// powerShellMkdirExistingNewCommand adds the Force parameter, which makes
// mkdir succeed when the directory already exists.
func powerShellMkdirExistingNewCommand(command *types.Command) ([]string, error) {
	if len(command.ScriptParts()) == 0 {
		return nil, nil
	}
	return []string{powerShellInsertArgument(command.Script, "-Force")}, nil
}

func init() {
	register(types.NewRule("powershell_mkdir_existing",
		powerShellMkdirExistingMatch,
		powerShellMkdirExistingNewCommand))
}
