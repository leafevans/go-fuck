package rules

import (
	"os"
	"strings"

	"github.com/leafevans/go-fuck/internal/types"
)

// powerShellCommandIs reports whether the command starts with one of the given
// PowerShell commands. PowerShell is case insensitive, so the comparison is too.
func powerShellCommandIs(command *types.Command, names ...string) bool {
	parts := command.ScriptParts()
	if len(parts) == 0 {
		return false
	}
	for _, name := range names {
		if strings.EqualFold(parts[0], name) {
			return true
		}
	}
	return false
}

// powerShellQuote quotes a value for the PowerShell command line.
func powerShellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "''") + "'"
}

// powerShellInsertArgument inserts argument after the first word of script.
func powerShellInsertArgument(script, argument string) string {
	trimmed := strings.TrimLeft(script, " \t")
	index := strings.IndexAny(trimmed, " \t")
	if index < 0 {
		return trimmed + " " + argument
	}
	return trimmed[:index] + " " + argument + trimmed[index:]
}

// powerShellParentDirectory returns the directory that has to be created for a
// path that does not exist, or an empty string when there is nothing to create.
//
// A path with a trailing separator names the directory itself; otherwise the
// parent directory is returned. Both Windows and POSIX separators are
// understood, because the path may come from a PowerShell error message or from
// the command line.
func powerShellParentDirectory(path string) string {
	trimmed := strings.TrimRight(path, "\\/")
	if trimmed == "" {
		return ""
	}
	if strings.HasSuffix(path, "\\") || strings.HasSuffix(path, "/") {
		return powerShellCreatableDirectory(trimmed)
	}
	index := strings.LastIndexAny(trimmed, "\\/")
	if index < 0 {
		return ""
	}
	return powerShellCreatableDirectory(trimmed[:index])
}

// powerShellCreatableDirectory drops directories that always exist, such as a
// drive root.
func powerShellCreatableDirectory(path string) string {
	if path == "" {
		return ""
	}
	if len(path) == 2 && path[1] == ':' {
		return ""
	}
	return path
}

// powerShellItemIsDirectory reports whether path exists and is a directory.
var powerShellItemIsDirectory = func(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// powerShellItemExists reports whether path exists.
var powerShellItemExists = func(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// powerShellDirectoryEntries returns the names in dir.
var powerShellDirectoryEntries = func(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	return names, nil
}
