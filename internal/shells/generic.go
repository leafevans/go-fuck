package shells

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/leafevans/go-fuck/internal/conf"
	"github.com/leafevans/go-fuck/internal/consts"
	"github.com/leafevans/go-fuck/internal/logs"
	"github.com/leafevans/go-fuck/internal/system"
)

// env returns the value of an environment variable.
func env(name string) string {
	return os.Getenv(name)
}

// processAncestorNames returns the executable names of the process ancestors,
// nearest first.
func processAncestorNames() []string {
	return system.AncestorNames()
}

// runOutput runs a command and returns its trimmed standard output. An empty
// string is returned when the command is unavailable or fails.
func runOutput(name string, args ...string) string {
	cmd := exec.Command(name, args...)
	cmd.Stdin = nil
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

// homePath expands a "~" prefixed path.
func homePath(path string) string {
	return system.ExpandHome(path)
}

// statHome stats a "~" prefixed path.
func statHome(path string) (os.FileInfo, error) {
	return os.Stat(homePath(path))
}

// Generic implements the behavior shared by all shells.
type Generic struct {
	name         string
	friendlyName string

	// historyFile returns the shell history file path, or "" when the shell
	// does not keep a readable history file.
	historyFile func() string
	// parseHistoryLine extracts the command from a single history file line.
	parseHistoryLine func(line string) string
	// versionFn returns the shell version, or "" when it cannot be detected.
	versionFn func() string

	historyOnce sync.Once
	history     []string
}

// NewGeneric returns the shell used when no other shell can be detected.
func NewGeneric() *Generic {
	return newGeneric("generic", "Generic Shell")
}

// newGeneric builds a Generic shell with the given name.
func newGeneric(name, friendlyName string) *Generic {
	return &Generic{
		name:             name,
		friendlyName:     friendlyName,
		historyFile:      func() string { return "" },
		parseHistoryLine: func(line string) string { return line },
		versionFn:        func() string { return "" },
	}
}

// Name returns the shell name.
func (g *Generic) Name() string {
	return g.name
}

// FriendlyName returns the human readable shell name.
func (g *Generic) FriendlyName() string {
	return g.friendlyName
}

// GetAliases returns the shell aliases. The generic shell has none.
func (g *Generic) GetAliases() map[string]string {
	return nil
}

// expandAliases replaces a leading alias with its value.
func (g *Generic) expandAliases(script string) string {
	aliases := g.GetAliases()
	if len(aliases) == 0 {
		return script
	}
	binary := strings.SplitN(script, " ", 2)[0]
	if replacement, ok := aliases[binary]; ok {
		return strings.Replace(script, binary, replacement, 1)
	}
	return script
}

// FromShell prepares a command before it is fixed.
func (g *Generic) FromShell(script string) string {
	return g.expandAliases(script)
}

// ToShell prepares a corrected command for execution.
func (g *Generic) ToShell(script string) string {
	return script
}

// AppAlias returns the alias definition for a POSIX compatible shell.
func (g *Generic) AppAlias(aliasName string) string {
	return "alias " + aliasName + "='eval " + dquote + "$(TF_ALIAS=" + aliasName +
		" " + consts.AppName + " " + dquote + "$(fc -ln -1)" + dquote + ")" + dquote + "'"
}

// InstantModeAlias returns the instant mode alias. The generic shell does not
// support instant mode, so a warning is printed and the regular alias is
// returned.
func (g *Generic) InstantModeAlias(aliasName string) string {
	logs.Warn("Instant mode not supported by your shell")
	return g.AppAlias(aliasName)
}

// And joins commands with "&&".
func (g *Generic) And(commands ...string) string {
	return strings.Join(commands, " && ")
}

// Or joins commands with "||".
func (g *Generic) Or(commands ...string) string {
	return strings.Join(commands, " || ")
}

// Quote returns a shell escaped version of s.
func (g *Generic) Quote(s string) string {
	return Quote(s)
}

// SplitCommand splits a command line using shell-like syntax.
func (g *Generic) SplitCommand(command string) []string {
	prepared := strings.ReplaceAll(command, backslash+" ", "??")
	parts, err := Split(prepared)
	if err != nil {
		return strings.Split(command, " ")
	}
	for i, part := range parts {
		parts[i] = strings.ReplaceAll(part, "??", backslash+" ")
	}
	return parts
}

// GetHistory returns the shell history, oldest entry first.
func (g *Generic) GetHistory() []string {
	g.historyOnce.Do(func() {
		g.history = g.readHistory()
	})
	return g.history
}

// readHistory reads and parses the shell history file.
func (g *Generic) readHistory() []string {
	path := g.historyFile()
	if path == "" {
		return nil
	}
	return readHistoryFile(path, g.parseHistoryLine)
}

// readHistoryFile parses a history file, applying the history_limit setting.
func readHistoryFile(path string, parse func(string) string) []string {
	file, err := os.Open(homePath(path))
	if err != nil {
		return nil
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil
	}

	if limit := conf.Current().HistoryLimit; limit > 0 && len(lines) > limit {
		lines = lines[len(lines)-limit:]
	}

	history := make([]string, 0, len(lines))
	for _, line := range lines {
		if prepared := strings.TrimSpace(parse(line)); prepared != "" {
			history = append(history, prepared)
		}
	}
	return history
}

// PutToHistory adds a command to the shell history. Most shells update their
// history on the shell side, so this is a no-op here.
func (g *Generic) PutToHistory(command string) error {
	return nil
}

// HowToConfigure describes how to configure the alias.
func (g *Generic) HowToConfigure() *Configuration {
	return nil
}

// Info returns the shell name and version.
func (g *Generic) Info() string {
	version := ""
	if g.versionFn != nil {
		version = safeVersion(g.versionFn)
	}
	return strings.TrimSpace(fmt.Sprintf("%s %s", g.friendlyName, version))
}

// safeVersion calls a version detection function and ignores failures.
func safeVersion(fn func() string) (version string) {
	defer func() {
		if recovered := recover(); recovered != nil {
			version = ""
		}
	}()
	return fn()
}

// BuiltinCommands returns the shell builtin command names.
func (g *Generic) BuiltinCommands() []string {
	return []string{
		"alias", "bg", "bind", "break", "builtin", "case", "cd", "command",
		"compgen", "complete", "continue", "declare", "dirs", "disown", "echo",
		"enable", "eval", "exec", "exit", "export", "fc", "fg", "getopts", "hash",
		"help", "history", "if", "jobs", "kill", "let", "local", "logout",
		"popd", "printf", "pushd", "pwd", "read", "readonly", "return", "set",
		"shift", "shopt", "source", "suspend", "test", "times", "trap", "type",
		"typeset", "ulimit", "umask", "unalias", "unset", "until", "wait",
		"while",
	}
}

// createConfiguration builds a shell configuration description.
func createConfiguration(content, path, reload string) *Configuration {
	_, err := os.Stat(homePath(path))
	return &Configuration{
		Content:                   content,
		Path:                      path,
		Reload:                    reload,
		CanConfigureAutomatically: err == nil,
	}
}

// parseAliasLine parses a "name<separator>value" alias line.
func parseAliasLine(line, separator string) (string, string, bool) {
	name, value, found := strings.Cut(line, separator)
	if !found {
		return "", "", false
	}
	name = strings.TrimSpace(name)
	value = strings.TrimSpace(value)
	if len(value) >= 2 {
		first, last := value[0], value[len(value)-1]
		if (first == charSingleQuote || first == charDoubleQuote) && first == last {
			value = value[1 : len(value)-1]
		}
	}
	if name == "" {
		return "", "", false
	}
	return name, value, true
}
