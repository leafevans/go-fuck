package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/v2/internal/types"
)

// powerShellPositionalParameterError matches the error that PowerShell prints
// when a command receives more positional arguments than it accepts, for
// example for "cd C:\Program Files".
const powerShellPositionalParameterError = "A positional parameter cannot be found that accepts argument"

// powerShellCDCommands are the PowerShell commands that change the directory.
var powerShellCDCommands = []string{"cd", "set-location", "sl", "chdir"}

// powerShellCDUnquotedPathMatch reports whether a directory change failed
// because the path contains unquoted spaces.
func powerShellCDUnquotedPathMatch(command *types.Command) bool {
	return powerShellCommandIs(command, powerShellCDCommands...) &&
		strings.Contains(command.Output, powerShellPositionalParameterError)
}

// powerShellCDUnquotedPathNewCommand quotes the path again.
//
// The raw part of the command is quoted instead of joining the shell split
// arguments, because splitting removes the backslashes of a Windows path.
func powerShellCDUnquotedPathNewCommand(command *types.Command) ([]string, error) {
	parts := command.ScriptParts()
	script := strings.TrimLeft(command.Script, " 	")
	index := strings.IndexAny(script, " 	")
	if len(parts) == 0 || index < 0 {
		return nil, nil
	}
	path := strings.TrimSpace(script[index:])
	if path == "" {
		return nil, nil
	}
	return []string{parts[0] + " " + powerShellQuote(path)}, nil
}

func init() {
	register(types.NewRule("powershell_cd_unquoted_path",
		powerShellCDUnquotedPathMatch,
		powerShellCDUnquotedPathNewCommand))
}
