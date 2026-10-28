// Package utils implements helpers shared by the rules: fuzzy matching,
// argument replacement, executable lookup and shell history access.
package utils

import (
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/leafevans/go-fuck/internal/conf"
	"github.com/leafevans/go-fuck/internal/shells"
	"github.com/leafevans/go-fuck/internal/system"
	"github.com/leafevans/go-fuck/internal/types"
)

// Which returns the absolute path of an executable, or an empty string when it
// is not available.
func Which(program string) string {
	return system.Which(program)
}

// GetAlias returns the alias name that invoked go-fuck.
func GetAlias() string {
	return conf.Alias()
}

// ResetCaches clears the process wide caches. It is intended for tests.
func ResetCaches() {
	executablesOnce = sync.Once{}
	executables = nil
}

var (
	executablesOnce sync.Once
	executables     []string
)

// GetAllExecutables returns the names of all executables that are available in
// PATH, together with the shell aliases.
func GetAllExecutables() []string {
	executablesOnce.Do(func() {
		executables = loadAllExecutables()
	})
	return executables
}

// loadAllExecutables scans PATH and the shell aliases.
func loadAllExecutables() []string {
	alias := GetAlias()
	entryPoints := map[string]bool{"go-fuck": true, "thefuck": true, "fuck": true}

	found := []string{}
	for _, path := range filepath.SplitList(os.Getenv("PATH")) {
		if !includePathInSearch(path) {
			continue
		}
		entries, err := os.ReadDir(path)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			name := entry.Name()
			base := strings.ToLower(trimExecutableSuffix(name))
			if entryPoints[base] || base == strings.ToLower(alias) {
				continue
			}
			found = append(found, name)
		}
	}

	for name := range shells.Current().GetAliases() {
		if name != alias {
			found = append(found, name)
		}
	}
	return found
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
	known := map[string]bool{}
	for _, name := range GetAllExecutables() {
		known[name] = true
	}
	for _, name := range shell.BuiltinCommands() {
		known[name] = true
	}

	valid := make([]string, 0, len(filtered))
	for _, line := range filtered {
		if strings.HasPrefix(line, alias) || line == command.Script {
			continue
		}
		executable := strings.SplitN(line, " ", 2)[0]
		if known[executable] {
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
	if index := strings.Index(script, search); index >= 0 {
		return script[:index] + " " + to + " " + script[index+len(search):]
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
	for _, name := range appNames {
		if base == name {
			return true
		}
	}
	return false
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

// SortedKeys returns the sorted keys of a string set.
func SortedKeys(set map[string]bool) []string {
	keys := make([]string, 0, len(set))
	for key := range set {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

// Itoa is a shorthand for strconv.Itoa.
func Itoa(value int) string {
	return strconv.Itoa(value)
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
