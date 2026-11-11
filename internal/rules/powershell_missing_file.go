package rules

import (
	"path/filepath"
	"regexp"
	"strings"

	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// powerShellMissingFilePatterns match the file path in the PowerShell errors
// for a file that does not exist.
var powerShellMissingFilePatterns = []*regexp.Regexp{
	regexp.MustCompile(`Cannot find path '([^']*)' because it does not exist\.`),
	regexp.MustCompile(`Cannot rename because item at '([^']*)' does not exist\.`),
}

// powerShellMissingFileCommands are the PowerShell commands that read, copy,
// move, remove or rename an existing item.
var powerShellMissingFileCommands = []string{
	"cp", "copy", "cpi", "copy-item",
	"mv", "move", "mi", "move-item",
	"rm", "del", "erase", "remove", "remove-item",
	"ren", "rename", "rename-item",
	"cat", "type", "gc", "get-content",
	"gi", "get-item",
}

// powerShellMissingFileMatch reports whether a PowerShell command failed
// because a file it names does not exist.
func powerShellMissingFileMatch(command *types.Command) bool {
	if !powerShellCommandIs(command, powerShellMissingFileCommands...) {
		return false
	}
	path, ok := powerShellMissingFile(command.Output)
	if !ok || path == "" || strings.ContainsAny(path, "*?") {
		return false
	}
	missing := filepath.Base(path)
	if missing == "" || powerShellArgumentIndex(command.Script, missing) < 0 {
		return false
	}
	return !powerShellItemExists(path)
}

// powerShellMissingFileNewCommand replaces the missing file with the closest
// existing entry in the same directory.
func powerShellMissingFileNewCommand(command *types.Command) ([]string, error) {
	path, ok := powerShellMissingFile(command.Output)
	if !ok {
		return nil, nil
	}
	missing := filepath.Base(path)
	if missing == "" || missing == "." || missing == string(filepath.Separator) {
		return nil, nil
	}
	index := powerShellArgumentIndex(command.Script, missing)
	if index < 0 {
		return nil, nil
	}
	names, err := powerShellDirectoryEntries(filepath.Dir(path))
	if err != nil {
		return nil, nil
	}
	closest := ""
	for _, name := range names {
		if strings.EqualFold(name, missing) {
			closest = name
			break
		}
	}
	if closest == "" {
		// PowerShell paths are case insensitive: compare lowercase names and
		// map the match back to the name as it exists on disk.
		lowered := make(map[string]string, len(names))
		loweredNames := make([]string, 0, len(names))
		for _, name := range names {
			lower := strings.ToLower(name)
			if _, ok := lowered[lower]; ok {
				continue
			}
			lowered[lower] = name
			loweredNames = append(loweredNames, lower)
		}
		candidate := utils.GetClosest(strings.ToLower(missing), loweredNames, 0.6, false)
		if candidate != "" {
			closest = lowered[candidate]
		}
	}
	if closest == "" || closest == missing {
		return nil, nil
	}
	return []string{command.Script[:index] + closest + command.Script[index+len(missing):]}, nil
}

// powerShellArgumentIndex returns the index of missing when it appears as a
// complete argument in script, or -1. The boundary check keeps a path such as
// config.txt from matching the missing file config.tx.
func powerShellArgumentIndex(script, missing string) int {
	pattern := regexp.MustCompile(`(^|[\s"'\\/(])` + regexp.QuoteMeta(missing) + `($|[\s"';|)])`)
	match := pattern.FindStringSubmatchIndex(script)
	if match == nil {
		return -1
	}
	return match[3]
}

// powerShellMissingFile returns the path from a matching PowerShell error.
func powerShellMissingFile(output string) (string, bool) {
	for _, pattern := range powerShellMissingFilePatterns {
		if match := pattern.FindStringSubmatch(output); match != nil {
			return match[1], true
		}
	}
	return "", false
}

func init() {
	register(types.NewRule("powershell_missing_file",
		powerShellMissingFileMatch,
		powerShellMissingFileNewCommand))
}
