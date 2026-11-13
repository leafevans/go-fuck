package rules

import (
	"regexp"
	"strings"

	"github.com/leafevans/go-fuck/v2/internal/types"
)

// powerShellCatDirectoryPath matches the path in the error PowerShell prints
// when Get-Content (cat, type, gc) is asked to read a directory.
var powerShellCatDirectoryPath = regexp.MustCompile(`Access to the path '([^']*)' is denied\.`)

// powerShellCatDirectoryCommands are the PowerShell commands that read a file.
var powerShellCatDirectoryCommands = []string{"cat", "type", "gc", "get-content"}

// powerShellCatDirectoryParsed returns the directory path from the error.
func powerShellCatDirectoryParsed(commandOutput string) (string, bool) {
	if match := powerShellCatDirectoryPath.FindStringSubmatch(commandOutput); match != nil {
		return match[1], true
	}
	return "", false
}

// powerShellCatDirectoryMatch reports whether a file reading command was used
// on a directory. Get-Content reports "Access to the path ... is denied." for
// a directory, which reads like a permission problem.
func powerShellCatDirectoryMatch(command *types.Command) bool {
	if !powerShellCommandIs(command, powerShellCatDirectoryCommands...) {
		return false
	}
	path, ok := powerShellCatDirectoryParsed(command.Output)
	if !ok || path == "" {
		return false
	}
	parts := command.ScriptParts()
	if len(parts) < 2 {
		return false
	}
	// The error names the resolved path; only match it when the command line
	// names the same directory. The path is split on both separators so that a
	// Windows path is understood on every platform.
	if !strings.EqualFold(powerShellPathBase(strings.Trim(parts[1], "\"'")), powerShellPathBase(path)) {
		return false
	}
	return powerShellItemIsDirectory(path)
}

// powerShellPathBase returns the last element of a path, accepting both the
// Windows and the POSIX separator on every platform.
func powerShellPathBase(value string) string {
	value = strings.ReplaceAll(value, "\\", "/")
	if index := strings.LastIndex(value, "/"); index >= 0 {
		return value[index+1:]
	}
	return value
}

// powerShellCatDirectoryNewCommand lists the directory instead of reading it.
func powerShellCatDirectoryNewCommand(command *types.Command) ([]string, error) {
	parts := command.ScriptParts()
	if len(parts) != 2 || strings.HasPrefix(parts[1], "-") {
		// Only the simple "command directory" form is translated; the
		// cmdlet parameter syntax is left alone.
		return nil, nil
	}
	return []string{"ls " + parts[1]}, nil
}

func init() {
	register(types.NewRule(
		"powershell_cat_directory",
		powerShellCatDirectoryMatch,
		powerShellCatDirectoryNewCommand,
	))
}
