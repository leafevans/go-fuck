// Package utils implements helpers shared by the rules: fuzzy matching,
// argument replacement, executable lookup and shell history access.
package utils

import (
	"path/filepath"
	"slices"
	"strings"
	"sync"

	"github.com/leafevans/go-fuck/v2/internal/conf"
	"github.com/leafevans/go-fuck/v2/internal/shells"
	"github.com/leafevans/go-fuck/v2/internal/types"
)

// GetAlias returns the alias name that invoked go-fuck.
func GetAlias() string {
	return conf.Alias()
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

// IsApp reports whether the command is a call to one of the given
// applications.
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
