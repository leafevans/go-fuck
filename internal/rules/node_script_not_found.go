package rules

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/leafevans/go-fuck/v4/internal/types"
	"github.com/leafevans/go-fuck/v4/internal/utils"
)

// nodeCannotFindModule matches the path of a module that Node.js cannot find.
var nodeCannotFindModule = newRegexp(`Cannot find module '([^']+)'`)

// nodeScriptExtensions are the file extensions that Node.js executes.
var nodeScriptExtensions = []string{".js", ".mjs", ".cjs", ".json", ".ts"}

// nodeDirectoryFiles returns the names of the files in dir.
var nodeDirectoryFiles = func(dir string) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		names = append(names, entry.Name())
	}
	return names
}

// nodePathBase returns the file name part of a Node.js path.
func nodePathBase(path string) string {
	if index := strings.LastIndexAny(path, "\\/"); index >= 0 {
		return path[index+1:]
	}
	return path
}

// nodePathDir returns the directory part of a Node.js path.
func nodePathDir(path string) string {
	if index := strings.LastIndexAny(path, "\\/"); index >= 0 {
		return path[:index]
	}
	return ""
}

// nodeAbsolutePath reports whether path is an absolute path.
func nodeAbsolutePath(path string) bool {
	if strings.HasPrefix(path, "/") {
		return true
	}
	return len(path) >= 3 && path[1] == ':' && (path[2] == '\\' || path[2] == '/')
}

// nodeScriptClosest returns the file that is closest to path, or an empty
// string.
func nodeScriptClosest(path string) string {
	base := nodePathBase(path)
	return utils.GetClosest(base, nodeDirectoryFiles(nodePathDir(path)), 0.6, false)
}

// nodeScriptName reports whether name has an extension that Node.js runs.
func nodeScriptName(name string) bool {
	lower := strings.ToLower(name)
	for _, extension := range nodeScriptExtensions {
		if strings.HasSuffix(lower, extension) {
			return true
		}
	}
	return false
}

// nodeScriptNotFoundMatch reports whether Node.js failed to load a missing
// file.
func nodeScriptNotFoundMatch(command *types.Command) bool {
	_, _, ok := nodeMissingEntryPoint(command)
	return ok
}

func nodeRuntimeOption(name string) commandOption {
	switch name {
	case "-r", "--require", "--import", "--env-file", "--env-file-if-exists":
		return valueOption
	case "--trace-warnings", "--no-warnings", "--enable-source-maps", "--no-deprecation":
		return flagOption
	default:
		return unknownOption
	}
}

func nodeMissingEntryPoint(command *types.Command) (commandWord, string, bool) {
	if !strings.Contains(command.Output, "MODULE_NOT_FOUND") || strings.Contains(command.Output, "Require stack:") {
		return commandWord{}, "", false
	}
	match := nodeCannotFindModule.FindStringSubmatch(command.Output)
	if match == nil || !nodeAbsolutePath(match[1]) || !nodeScriptName(nodePathBase(match[1])) {
		return commandWord{}, "", false
	}
	words := simpleCommandWords(command.Script)
	if len(words) < 2 || !strings.EqualFold(words[0].text, "node") {
		return commandWord{}, "", false
	}
	for i := 1; i < len(words); {
		if words[i].text == "--" {
			i++
			if i == len(words) {
				break
			}
		} else if strings.HasPrefix(words[i].text, "-") {
			end, ok := optionEnd(words, i, nodeRuntimeOption)
			if !ok {
				break
			}
			i = end
			continue
		}
		word := words[i]
		entry, err := filepath.Abs(word.text)
		if err != nil || entry != filepath.Clean(match[1]) {
			break
		}
		if closest := nodeScriptClosest(match[1]); closest != "" {
			prefix := word.text[:len(word.text)-len(nodePathBase(word.text))]
			if prefix == "" && strings.HasPrefix(closest, "-") {
				prefix = "./"
			}
			return word, prefix + closest, true
		}
		break
	}
	return commandWord{}, "", false
}

// nodeScriptNotFoundNewCommand replaces only the entry point, not a missing
// dependency or a similarly named script argument.
func nodeScriptNotFoundNewCommand(command *types.Command) ([]string, error) {
	word, replacement, ok := nodeMissingEntryPoint(command)
	if !ok {
		return nil, nil
	}
	return replaceCommandWord(command.Script, word, "", replacement), nil
}

func init() {
	register(types.NewRule("node_script_not_found",
		nodeScriptNotFoundMatch,
		nodeScriptNotFoundNewCommand))
}
