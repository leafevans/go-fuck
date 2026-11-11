package rules

import (
	"os"
	"strings"
)

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

// cmdSplit splits a cmd.exe command into arguments. Double quotes group
// arguments and are removed; backslashes are kept because they are path
// separators in cmd.exe.
func cmdSplit(script string) []string {
	var parts []string
	var current strings.Builder
	quoted := false
	for _, character := range script {
		switch {
		case character == '"':
			quoted = !quoted
		case !quoted && (character == ' ' || character == '\t'):
			if current.Len() > 0 {
				parts = append(parts, current.String())
				current.Reset()
			}
		default:
			current.WriteRune(character)
		}
	}
	if current.Len() > 0 {
		parts = append(parts, current.String())
	}
	return parts
}

// cmdParentDirectory returns the directory that has to be created for a cmd.exe
// path, or an empty string when there is nothing to create.
func cmdParentDirectory(path string) string {
	trimmed := strings.TrimRight(path, "\\/")
	if trimmed == "" {
		return ""
	}
	if strings.HasSuffix(path, "\\") || strings.HasSuffix(path, "/") {
		return cmdCreatableDirectory(trimmed)
	}
	index := strings.LastIndexAny(trimmed, "\\/")
	if index < 0 {
		return ""
	}
	return cmdCreatableDirectory(trimmed[:index])
}

// cmdCreatableDirectory drops paths that always exist, such as a drive root.
func cmdCreatableDirectory(path string) string {
	if path == "" || (len(path) == 2 && path[1] == ':') {
		return ""
	}
	return path
}

// cmdDirectoryMissing reports whether path does not exist.
var cmdDirectoryMissing = func(path string) bool {
	_, err := os.Stat(path)
	return err != nil
}
