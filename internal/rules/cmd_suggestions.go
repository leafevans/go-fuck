package rules

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/leafevans/go-fuck/internal/utils"
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

// cmdFirstRawArgument returns the first argument after the command name, both
// as it appears in the script (including quotes) and as a plain value. Flags
// are skipped.
func cmdFirstRawArgument(script string) (raw, value string) {
	index := 0
	for index < len(script) && script[index] != ' ' && script[index] != '\t' {
		index++
	}
	for index < len(script) {
		for index < len(script) && (script[index] == ' ' || script[index] == '\t') {
			index++
		}
		if index >= len(script) {
			return "", ""
		}
		start := index
		if script[index] == '"' {
			index++
			for index < len(script) && script[index] != '"' {
				index++
			}
			if index < len(script) {
				index++
			}
			raw = script[start:index]
			return raw, strings.Trim(raw, "\"")
		}
		for index < len(script) && script[index] != ' ' && script[index] != '\t' {
			index++
		}
		raw = script[start:index]
		if strings.HasPrefix(raw, "/") || strings.HasPrefix(raw, "-") || strings.HasPrefix(raw, "+") {
			continue
		}
		return raw, raw
	}
	return "", ""
}

// cmdFileExists reports whether path exists.
var cmdFileExists = func(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// cmdDirectoryEntries returns the names in dir.
var cmdDirectoryEntries = func(dir string) ([]string, error) {
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

// cmdClosestFile returns the existing file or directory in the argument tree
// that is closest to argument.
func cmdClosestFile(argument string) string {
	dir := filepath.Dir(argument)
	names, err := cmdDirectoryEntries(dir)
	if err != nil {
		return ""
	}
	base := filepath.Base(argument)
	// Windows paths are case insensitive, so an exact match with a different
	// case is the best correction even when the similarity is low.
	for _, name := range names {
		if strings.EqualFold(name, base) {
			if dir == "." {
				return name
			}
			return filepath.Join(dir, name)
		}
	}
	closest := utils.GetClosest(base, names, 0.6, false)
	if closest == "" {
		return ""
	}
	if dir == "." {
		return closest
	}
	return filepath.Join(dir, closest)
}

// cmdReplaceRawArgument replaces the raw argument with value in the script.
func cmdReplaceRawArgument(script, raw, value string) string {
	if strings.ContainsAny(value, " \t") {
		value = cmdQuote(value)
	}
	return strings.Replace(script, raw, value, 1)
}

// cmdFirstWord returns the first word of a cmd.exe command without
// allocating.
func cmdFirstWord(script string) string {
	trimmed := strings.TrimLeft(script, " \t")
	if index := strings.IndexAny(trimmed, " \t"); index >= 0 {
		return trimmed[:index]
	}
	return trimmed
}

// cmdPathIsDirectory reports whether path exists and is a directory.
var cmdPathIsDirectory = func(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
