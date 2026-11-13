package rules

import (
	"regexp"
	"strings"

	"github.com/leafevans/go-fuck/v2/internal/shells"
	"github.com/leafevans/go-fuck/v2/internal/types"
)

// powerShellMissingPath matches the destination path of a failed copy.
var powerShellMissingPath = regexp.MustCompile(`Could not find a part of the path '([^']*)'`)

// powerShellMissingDirectoryCommands are the PowerShell commands that write a
// file, copy an item or move an item.
var powerShellMissingDirectoryCommands = []string{
	"cp", "copy", "cpi", "copy-item",
	"mv", "move", "mi", "move-item",
	"set-content", "add-content", "out-file", "new-item",
}

// powerShellMissingDirectoryMatch reports whether a command failed because a
// directory of the destination does not exist.
func powerShellMissingDirectoryMatch(command *types.Command) bool {
	return powerShellCommandIs(command, powerShellMissingDirectoryCommands...) &&
		strings.Contains(command.Output, "Could not find a part of the path")
}

// powerShellMissingDirectoryNewCommand creates the missing directory and
// retries the command.
func powerShellMissingDirectoryNewCommand(command *types.Command) ([]string, error) {
	destination := ""
	if match := powerShellMissingPath.FindStringSubmatch(command.Output); match != nil {
		destination = match[1]
	} else {
		destination = powerShellLastArgument(command.Script)
	}
	directory := powerShellParentDirectory(destination)
	if directory == "" {
		return nil, nil
	}
	create := "mkdir -Force " + powerShellQuote(directory)
	return []string{shells.Current().And(create, command.Script)}, nil
}

// powerShellLastArgument returns the last argument that is not a flag.
//
// The raw command is used instead of the shell split arguments, because
// splitting removes the backslashes of a Windows path.
func powerShellLastArgument(script string) string {
	tokens := strings.Fields(script)
	for index := len(tokens) - 1; index > 0; index-- {
		if !strings.HasPrefix(tokens[index], "-") {
			return tokens[index]
		}
	}
	return ""
}

func init() {
	register(types.NewRule("powershell_missing_directory",
		powerShellMissingDirectoryMatch,
		powerShellMissingDirectoryNewCommand))
}
