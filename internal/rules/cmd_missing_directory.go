package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/v2/internal/types"
)

// cmdMissingPathError is the cmd.exe error for a missing path component.
const cmdMissingPathError = "The system cannot find the path specified."

// cmdMissingDirectoryMatch reports whether copy or move failed because a
// directory of the destination is missing.
func cmdMissingDirectoryMatch(command *types.Command) bool {
	parts := cmdSplit(command.Script)
	if len(parts) < 3 {
		return false
	}
	if !strings.EqualFold(parts[0], "copy") && !strings.EqualFold(parts[0], "move") {
		return false
	}
	if !strings.Contains(command.Output, cmdMissingPathError) {
		return false
	}
	destination := cmdCopyDestination(parts)
	if destination == "" {
		return false
	}
	directory := cmdParentDirectory(destination)
	return directory != "" && !strings.HasPrefix(directory, "\\\\") && cmdDirectoryMissing(directory)
}

// cmdMissingDirectoryNewCommand creates the missing directory and retries.
func cmdMissingDirectoryNewCommand(command *types.Command) ([]string, error) {
	parts := cmdSplit(command.Script)
	if len(parts) < 3 {
		return nil, nil
	}
	destination := cmdCopyDestination(parts)
	if destination == "" {
		return nil, nil
	}
	directory := cmdParentDirectory(destination)
	if directory == "" {
		return nil, nil
	}
	return []string{"mkdir " + cmdQuote(directory) + " && " + command.Script}, nil
}

// cmdCopyDestination returns the destination argument of copy or move.
func cmdCopyDestination(parts []string) string {
	for index := len(parts) - 1; index >= 2; index-- {
		if strings.HasPrefix(parts[index], "/") {
			continue
		}
		return parts[index]
	}
	return ""
}

func init() {
	register(types.NewRule("cmd_missing_directory",
		cmdMissingDirectoryMatch,
		cmdMissingDirectoryNewCommand))
}
