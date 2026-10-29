// Package utils implements helpers shared by the rules: fuzzy matching,
// argument replacement, executable lookup and shell history access.
package utils

import (
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"sync"

	"github.com/leafevans/go-fuck/internal/conf"
	"github.com/leafevans/go-fuck/internal/shells"
	"github.com/leafevans/go-fuck/internal/system"
	"github.com/leafevans/go-fuck/internal/types"
)

// whichCacheEntry is a memoized executable lookup. The entry is only valid
// while PATH is unchanged, so tests that modify PATH keep working.
type whichCacheEntry struct {
	path    string
	pathEnv string
}

var whichCache sync.Map

// Which returns the absolute path of an executable, or an empty string when it
// is not available.
//
// The lookup is memoized because rule matching asks for the same programs many
// times and every miss costs a number of file system probes on Windows.
func Which(program string) string {
	pathEnv := os.Getenv("PATH")
	if cached, ok := whichCache.Load(program); ok {
		entry := cached.(whichCacheEntry)
		if entry.pathEnv == pathEnv {
			return entry.path
		}
	}

	resolved := system.Which(program)
	whichCache.Store(program, whichCacheEntry{path: resolved, pathEnv: pathEnv})
	return resolved
}

// GetAlias returns the alias name that invoked go-fuck.
func GetAlias() string {
	return conf.Alias()
}

// ResetCaches clears the process wide caches. It is intended for tests.
func ResetCaches() {
	executablesOnce = sync.Once{}
	executables = nil
	executableSet = nil
	whichCache = sync.Map{}
}

var (
	executablesOnce sync.Once
	executables     []string
	executableSet   map[string]bool
)

// GetAllExecutables returns the names of all executables that are available in
// PATH, together with the shell aliases.
func GetAllExecutables() []string {
	executablesOnce.Do(func() {
		executables = loadAllExecutables()
		executableSet = make(map[string]bool, len(executables))
		for _, name := range executables {
			executableSet[name] = true
		}
	})
	return executables
}

// IsExecutable reports whether name is available in PATH or as a shell alias.
func IsExecutable(name string) bool {
	GetAllExecutables()
	return executableSet[name]
}

// executableEntryPoints are the go-fuck binaries that must not be suggested as
// corrected commands.
var executableEntryPoints = map[string]bool{
	"go-fuck": true,
	"thefuck": true,
	"fuck":    true,
}

// loadAllExecutables scans PATH and the shell aliases.
//
// The PATH directories are scanned in parallel because a typical Windows PATH
// contains many directories and the scan is latency bound.
func loadAllExecutables() []string {
	alias := GetAlias()
	aliasLower := strings.ToLower(alias)

	paths := filepath.SplitList(os.Getenv("PATH"))
	dirs := make([]string, 0, len(paths))
	for _, path := range paths {
		if includePathInSearch(path) {
			dirs = append(dirs, path)
		}
	}

	perDir := make([][]string, len(dirs))
	if len(dirs) > 0 {
		workers := min(runtime.NumCPU(), len(dirs))
		jobs := make(chan int)
		var waitGroup sync.WaitGroup
		waitGroup.Add(workers)
		for range workers {
			go func() {
				defer waitGroup.Done()
				for index := range jobs {
					perDir[index] = readExecutableDir(dirs[index], aliasLower)
				}
			}()
		}
		for index := range dirs {
			jobs <- index
		}
		close(jobs)
		waitGroup.Wait()
	}

	found := []string{}
	for _, names := range perDir {
		found = append(found, names...)
	}

	for name := range shells.Current().GetAliases() {
		if name != alias {
			found = append(found, name)
		}
	}
	return found
}

// readExecutableDir returns the names of the executable files of a PATH
// directory. The result order matches os.ReadDir, so the flattened list is
// deterministic and follows PATH order.
func readExecutableDir(dir, aliasLower string) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		base := strings.ToLower(trimExecutableSuffix(name))
		if executableEntryPoints[base] || base == aliasLower {
			continue
		}
		names = append(names, name)
	}
	return names
}

// trimExecutableSuffix removes a Windows executable suffix.
func trimExecutableSuffix(name string) string {
	if len(name) > 4 && strings.EqualFold(name[len(name)-4:], ".exe") {
		return name[:len(name)-4]
	}
	return name
}

// includePathInSearch reports whether a PATH entry should be searched.
func includePathInSearch(path string) bool {
	for _, prefix := range conf.Current().ExcludedSearchPathPrefixes {
		if prefix != "" && strings.HasPrefix(path, prefix) {
			return false
		}
	}
	return true
}

// GetValidHistoryWithoutCurrent returns the history entries that can be
// suggested as corrected commands.
func GetValidHistoryWithoutCurrent(command *types.Command) []string {
	history := shells.Current().GetHistory()
	alias := GetAlias()

	filtered := make([]string, 0, len(history))
	previous := ""
	hasPrevious := false
	for _, line := range history {
		if hasPrevious && line != alias {
			filtered = append(filtered, previous)
		}
		previous = line
		hasPrevious = true
	}
	if hasPrevious {
		filtered = append(filtered, previous)
	}

	shell := shells.Current()
	// GetAllExecutables populates the cached executable set as a side effect.
	GetAllExecutables()
	builtins := make(map[string]bool, len(shell.BuiltinCommands()))
	for _, name := range shell.BuiltinCommands() {
		builtins[name] = true
	}

	valid := make([]string, 0, len(filtered))
	for _, line := range filtered {
		if strings.HasPrefix(line, alias) || line == command.Script {
			continue
		}
		executable := strings.SplitN(line, " ", 2)[0]
		if executableSet[executable] || builtins[executable] {
			valid = append(valid, line)
		}
	}
	return valid
}

// ReplaceArgument replaces a single command line argument.
func ReplaceArgument(script, from, to string) string {
	suffix := " " + from
	if strings.HasSuffix(script, suffix) {
		return script[:len(script)-len(suffix)] + " " + to
	}
	search := " " + from + " "
	if before, after, ok := strings.Cut(script, search); ok {
		return before + " " + to + " " + after
	}
	return script
}

// ReplaceCommand replaces a misspelled command with its closest matches.
func ReplaceCommand(command *types.Command, broken string, matched []string) []string {
	closeMatches := GetCloseMatches(broken, matched, 0, 0.1)
	results := make([]string, 0, len(closeMatches))
	for _, match := range closeMatches {
		results = append(results, ReplaceArgument(command.Script, broken, strings.TrimSpace(match)))
	}
	return results
}

// GetAllMatchedCommands extracts the suggested commands from a command output
// that lists suggestions after one of the given separators.
func GetAllMatchedCommands(commandOutput string, separators ...string) []string {
	results := []string{}
	shouldYield := false
	for _, line := range strings.Split(commandOutput, "\n") {
		matched := false
		for _, separator := range separators {
			if strings.Contains(line, separator) {
				shouldYield = true
				matched = true
				break
			}
		}
		if matched {
			continue
		}
		if shouldYield && line != "" {
			results = append(results, strings.TrimSpace(line))
		}
	}
	return results
}

// IsApp reports whether the command is a call to one of the given applications.
//
// atLeast is the minimum number of additional arguments that the command must
// have.
func IsApp(command *types.Command, atLeast int, appNames ...string) bool {
	parts := command.ScriptParts()
	if len(parts) <= atLeast {
		return false
	}
	base := filepath.Base(parts[0])
	return slices.Contains(appNames, base)
}

// ForApp wraps a function so that it is only called for the given
// applications.
func ForApp[T any](appNames []string, atLeast int, fn func(*types.Command) T) func(*types.Command) T {
	return func(command *types.Command) T {
		if IsApp(command, atLeast, appNames...) {
			return fn(command)
		}
		var zero T
		return zero
	}
}

// Memoize caches calls to a function without arguments.
func Memoize[T any](fn func() T) func() T {
	var (
		once  sync.Once
		value T
	)
	return func() T {
		once.Do(func() {
			value = fn()
		})
		return value
	}
}

// Memoize1 caches calls to a function with a single comparable argument.
func Memoize1[K comparable, V any](fn func(K) V) func(K) V {
	var cache sync.Map
	return func(key K) V {
		if value, ok := cache.Load(key); ok {
			return value.(V)
		}
		value := fn(key)
		cache.Store(key, value)
		return value
	}
}

// Contains reports whether values contains value.
func Contains[T comparable](values []T, value T) bool {
	return slices.Contains(values, value)
}

// Index returns the index of the first occurrence of value in values, or -1.
func Index[T comparable](values []T, value T) int {
	return slices.Index(values, value)
}

// ForAppNew wraps a correction function so that it is only called for the
// given applications.
func ForAppNew(appNames []string, atLeast int, fn func(*types.Command) ([]string, error)) func(*types.Command) ([]string, error) {
	return func(command *types.Command) ([]string, error) {
		if IsApp(command, atLeast, appNames...) {
			return fn(command)
		}
		return nil, nil
	}
}
