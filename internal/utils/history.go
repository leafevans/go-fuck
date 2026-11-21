package utils

import (
	"strings"

	"github.com/leafevans/go-fuck/v2/internal/shells"
	"github.com/leafevans/go-fuck/v2/internal/types"
)

// GetValidHistoryWithoutCurrent returns the history entries that can be
// suggested as corrected commands.
func GetValidHistoryWithoutCurrent(command *types.Command) []string {
	shell := shells.Current()
	history := shell.GetHistory()
	alias := GetAlias()

	// GetAllExecutables populates the cached executable set as a side effect.
	GetAllExecutables()
	commands := shell.BuiltinCommands()
	builtins := make(map[string]bool, len(commands))
	for _, name := range commands {
		builtins[name] = true
	}

	valid := make([]string, 0, len(history))
	for i, line := range history {
		executable := HistoryExecutable(line)
		if executable == alias || line == command.Script {
			continue
		}
		// An alias invocation marks the preceding entry as a failed command.
		// Compare complete names so a short alias such as f does not hide fd.
		if i+1 < len(history) && HistoryExecutable(history[i+1]) == alias {
			continue
		}
		if executableSet[executable] || builtins[executable] {
			valid = append(valid, line)
		}
	}
	return valid
}

// HistoryExecutable reads the leading name without allocating argument slices.
func HistoryExecutable(line string) string {
	line = strings.TrimLeft(line, " \t\r\n")
	if end := strings.IndexAny(line, " \t\r\n"); end >= 0 {
		return line[:end]
	}
	return line
}
