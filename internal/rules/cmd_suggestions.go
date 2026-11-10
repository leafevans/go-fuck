package rules

import "strings"

// cmdInsertArgument inserts argument after the first word of a cmd.exe
// command.
func cmdInsertArgument(script, argument string) string {
	trimmed := strings.TrimLeft(script, " \t")
	index := strings.IndexAny(trimmed, " \t")
	if index < 0 {
		return trimmed + " " + argument
	}
	return trimmed[:index] + " " + argument + trimmed[index:]
}

// cmdQuote quotes a value for the cmd.exe command line.
func cmdQuote(value string) string {
	return "\"" + value + "\""
}
