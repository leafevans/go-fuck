package rules

import (
	"os"
	"regexp"
	"strings"

	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// nodeCannotFindModule matches the path of a module that Node.js cannot find.
var nodeCannotFindModule = regexp.MustCompile(`Cannot find module '([^']+)'`)

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
	parts := command.ScriptParts()
	if len(parts) == 0 || !strings.EqualFold(parts[0], "node") {
		return false
	}
	if !strings.Contains(command.Output, "MODULE_NOT_FOUND") {
		return false
	}
	match := nodeCannotFindModule.FindStringSubmatch(command.Output)
	if match == nil {
		return false
	}
	path := match[1]
	if !nodeAbsolutePath(path) {
		return false
	}
	base := nodePathBase(path)
	if !nodeScriptName(base) {
		return false
	}
	return nodeScriptClosest(path) != ""
}

// nodeScriptNotFoundNewCommand replaces the missing file with the closest one.
func nodeScriptNotFoundNewCommand(command *types.Command) ([]string, error) {
	match := nodeCannotFindModule.FindStringSubmatch(command.Output)
	if match == nil {
		return nil, nil
	}
	path := match[1]
	base := nodePathBase(path)
	closest := nodeScriptClosest(path)
	if closest == "" {
		return nil, nil
	}
	fixed := utils.ReplaceArgument(command.Script, base, closest)
	if fixed == command.Script {
		fixed = strings.Replace(command.Script, path, nodePathDir(path)+"/"+closest, 1)
	}
	if fixed == "" || fixed == command.Script {
		return nil, nil
	}
	return []string{fixed}, nil
}

func init() {
	register(types.NewRule("node_script_not_found",
		nodeScriptNotFoundMatch,
		nodeScriptNotFoundNewCommand))
}
