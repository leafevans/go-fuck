package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/internal/types"
)

// cmdPushdPathNotFound is the cmd.exe error for a missing pushd path.
const cmdPushdPathNotFound = "The system cannot find the path specified."

// cmdPushdDirectory returns the directory argument of a pushd command.
func cmdPushdDirectory(command *types.Command) string {
	trimmed := strings.TrimLeft(command.Script, " \t")
	index := strings.IndexAny(trimmed, " \t")
	if index < 0 {
		return ""
	}
	rest := strings.TrimSpace(trimmed[index:])
	if len(rest) >= 2 && rest[0] == '"' && rest[len(rest)-1] == '"' {
		rest = rest[1 : len(rest)-1]
	}
	return rest
}

// cmdPushdMkdirMatch reports whether pushd failed because the directory is
// missing.
func cmdPushdMkdirMatch(command *types.Command) bool {
	if !strings.EqualFold(cmdFirstWord(command.Script), "pushd") {
		return false
	}
	if !strings.Contains(command.Output, cmdPushdPathNotFound) {
		return false
	}
	directory := cmdPushdDirectory(command)
	return directory != "" && !strings.HasPrefix(directory, "\\\\")
}

// cmdPushdMkdirNewCommand creates the directory and changes into it.
func cmdPushdMkdirNewCommand(command *types.Command) ([]string, error) {
	directory := cmdPushdDirectory(command)
	if directory == "" {
		return nil, nil
	}
	quoted := cmdQuote(directory)
	return []string{"mkdir " + quoted + " && pushd " + quoted}, nil
}

func init() {
	register(types.NewRule("cmd_pushd_mkdir",
		cmdPushdMkdirMatch,
		cmdPushdMkdirNewCommand))
}
