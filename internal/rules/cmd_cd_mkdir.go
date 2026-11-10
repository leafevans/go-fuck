package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/internal/types"
)

// cmdPathNotFoundError is the error that cmd.exe prints for a missing path.
const cmdPathNotFoundError = "The system cannot find the path specified."

// cmdCDDirectory returns the directory argument of a cd command.
//
// The raw command is used instead of the shell split arguments, because
// splitting removes the backslashes of a Windows path.
func cmdCDDirectory(command *types.Command) string {
	trimmed := strings.TrimLeft(command.Script, " \t")
	index := strings.IndexAny(trimmed, " \t")
	if index < 0 {
		return ""
	}
	rest := strings.TrimSpace(trimmed[index:])
	if len(rest) >= 2 && (rest[1] == 'd' || rest[1] == 'D') && rest[0] == '/' {
		rest = strings.TrimSpace(rest[2:])
	}
	if len(rest) >= 2 && rest[0] == '"' && rest[len(rest)-1] == '"' {
		rest = rest[1 : len(rest)-1]
	}
	return rest
}

// cmdCDMkdirMatch reports whether cd failed because the directory is missing.
func cmdCDMkdirMatch(command *types.Command) bool {
	parts := command.ScriptParts()
	if len(parts) == 0 {
		return false
	}
	if !strings.EqualFold(parts[0], "cd") && !strings.EqualFold(parts[0], "chdir") {
		return false
	}
	if !strings.Contains(command.Output, cmdPathNotFoundError) {
		return false
	}
	return cmdCDDirectory(command) != ""
}

// cmdCDMkdirNewCommand creates the directory and changes into it.
func cmdCDMkdirNewCommand(command *types.Command) ([]string, error) {
	directory := cmdCDDirectory(command)
	if directory == "" {
		return nil, nil
	}
	quoted := cmdQuote(directory)
	return []string{"mkdir " + quoted + " && cd /d " + quoted}, nil
}

func init() {
	register(types.NewRule("cmd_cd_mkdir",
		cmdCDMkdirMatch,
		cmdCDMkdirNewCommand))
}
