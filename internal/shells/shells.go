// Package shells implements the shell specific behavior of go-fuck: alias
// generation, alias expansion, history access and command quoting.
package shells

import (
	"path/filepath"
	"sync"
)

// Configuration describes how to install the alias in a shell configuration
// file.
type Configuration struct {
	// Content is the line that has to be added to the configuration file.
	Content string
	// Path is the configuration file path, with a "~" prefix when relative to
	// the home directory.
	Path string
	// Reload is the command that reloads the configuration file.
	Reload string
	// CanConfigureAutomatically reports whether go-fuck may append the alias to
	// the configuration file without user interaction.
	CanConfigureAutomatically bool
}

// Shell abstracts the shell specific behavior.
type Shell interface {
	// Name returns the shell name used in TF_SHELL and in the settings.
	Name() string
	// FriendlyName returns the human readable shell name.
	FriendlyName() string
	// AppAlias returns the shell code that defines the alias.
	AppAlias(aliasName string) string
	// InstantModeAlias returns the alias used by the experimental instant mode.
	InstantModeAlias(aliasName string) string
	// GetAliases returns alias name to value mapping.
	GetAliases() map[string]string
	// FromShell prepares a command before it is fixed, for example by
	// expanding shell aliases.
	FromShell(command string) string
	// ToShell prepares a corrected command for execution in the shell.
	ToShell(command string) string
	// And joins commands with the shell's AND operator.
	And(commands ...string) string
	// Or joins commands with the shell's OR operator.
	Or(commands ...string) string
	// Quote returns a shell escaped version of s.
	Quote(s string) string
	// SplitCommand splits a command line using shell-like syntax.
	SplitCommand(command string) []string
	// GetHistory returns the shell history, oldest entry first.
	GetHistory() []string
	// PutToHistory adds a command to the shell history.
	PutToHistory(command string) error
	// HowToConfigure describes how to configure the alias for this shell.
	HowToConfigure() *Configuration
	// Info returns the shell name and version.
	Info() string
	// BuiltinCommands returns the shell builtin command names.
	BuiltinCommands() []string
}

// registry maps shell executable names to constructors.
var registry = map[string]func() Shell{
	"bash":       func() Shell { return NewBash() },
	"zsh":        func() Shell { return NewZsh() },
	"fish":       func() Shell { return NewFish() },
	"csh":        func() Shell { return NewTcsh() },
	"tcsh":       func() Shell { return NewTcsh() },
	"powershell": func() Shell { return NewPowershell() },
	"pwsh":       func() Shell { return NewPowershell() },
}

var (
	currentMu sync.RWMutex
	current   Shell = NewGeneric()
)

// Current returns the shell that invoked go-fuck.
func Current() Shell {
	currentMu.RLock()
	defer currentMu.RUnlock()
	return current
}

// Set replaces the current shell. It is intended for tests and for the
// detector.
func Set(s Shell) {
	currentMu.Lock()
	defer currentMu.Unlock()
	current = s
}

// FromEnv returns the shell declared in TF_SHELL, or nil.
func FromEnv() Shell {
	name := env("TF_SHELL")
	if constructor, ok := registry[name]; ok {
		return constructor()
	}
	return nil
}

// Detect returns the shell that invoked go-fuck.
//
// The shell reported by the alias through TF_SHELL takes precedence, because
// process names are not always reliable (for example inside containers or when
// a wrapper such as sudo is used). When TF_SHELL is not set the process
// ancestors are inspected.
func Detect() Shell {
	if shell := FromEnv(); shell != nil {
		return shell
	}
	for _, name := range processAncestorNames() {
		base := filepath.Base(name)
		base = trimExeSuffix(base)
		if constructor, ok := registry[base]; ok {
			return constructor()
		}
	}
	return NewGeneric()
}

// trimExeSuffix removes a Windows executable suffix from a process name.
func trimExeSuffix(name string) string {
	if len(name) > 4 && (name[len(name)-4:] == ".exe" || name[len(name)-4:] == ".EXE") {
		return name[:len(name)-4]
	}
	return name
}
