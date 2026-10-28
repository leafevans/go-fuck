package shells

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/leafevans/go-fuck/internal/conf"
	"github.com/leafevans/go-fuck/internal/consts"
	"github.com/leafevans/go-fuck/internal/logs"
)

// fishAliasTemplate defines the alias function for fish.
const fishAliasTemplate = `function %[1]s -d "Correct your previous console command"
  set -l fucked_up_command $history[1]
  env TF_SHELL=fish TF_ALIAS=%[1]s %[2]s $fucked_up_command %[3]s $argv | read -l unfucked_command
  if [ "$unfucked_command" != "" ]
    eval $unfucked_command
%[4]s
  end
end`

// fishAlterHistory removes the failed command from the fish history.
const fishAlterHistory = `    builtin history delete --exact --case-sensitive -- $fucked_up_command
    builtin history merge`

// Fish implements the fish shell specific behavior.
type Fish struct {
	*Generic

	aliasesOnce sync.Once
	aliases     map[string]string
}

// NewFish returns a fish shell.
func NewFish() *Fish {
	generic := newGeneric("fish", "Fish Shell")
	generic.historyFile = fishHistoryFile
	generic.parseHistoryLine = fishParseHistoryLine
	generic.versionFn = fishVersion
	return &Fish{Generic: generic}
}

// AppAlias returns the fish alias function.
func (f *Fish) AppAlias(aliasName string) string {
	alterHistory := ""
	if conf.Current().AlterHistory {
		alterHistory = fishAlterHistory
	}
	return fmt.Sprintf(fishAliasTemplate, aliasName, consts.AppName,
		consts.ArgumentPlaceholder, alterHistory)
}

// InstantModeAlias falls back to the regular alias because instant mode is not
// supported.
func (f *Fish) InstantModeAlias(aliasName string) string {
	logs.Warn("Instant mode not supported by go-fuck")
	return f.AppAlias(aliasName)
}

// And joins commands with the fish AND operator.
func (f *Fish) And(commands ...string) string {
	return strings.Join(commands, "; and ")
}

// Or joins commands with the fish OR operator.
func (f *Fish) Or(commands ...string) string {
	return strings.Join(commands, "; or ")
}

// GetAliases returns the fish functions and aliases.
func (f *Fish) GetAliases() map[string]string {
	f.aliasesOnce.Do(func() {
		f.aliases = f.loadAliases()
	})
	return f.aliases
}

// loadAliases queries fish for its functions and aliases.
func (f *Fish) loadAliases() map[string]string {
	overridden := f.overriddenAliases()
	aliases := map[string]string{}

	for _, name := range strings.Split(runOutput("fish", "-ic", "functions"), newline) {
		name = strings.TrimSpace(name)
		if name == "" || overridden[name] {
			continue
		}
		aliases[name] = name
	}

	for _, line := range strings.Split(runOutput("fish", "-ic", "alias"), newline) {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		line = strings.TrimPrefix(line, "alias ")
		var name, value string
		if parts := strings.SplitN(line, " ", 2); len(parts) == 2 {
			name, value = parts[0], parts[1]
		} else if parts := strings.SplitN(line, "=", 2); len(parts) == 2 {
			name, value = parts[0], parts[1]
		} else {
			continue
		}
		if overridden[name] {
			continue
		}
		aliases[name] = value
	}
	return aliases
}

// overriddenAliases returns the set of names that must not be expanded.
func (f *Fish) overriddenAliases() map[string]bool {
	raw := env("THEFUCK_OVERRIDDEN_ALIASES")
	if raw == "" {
		raw = env("TF_OVERRIDDEN_ALIASES")
	}
	overridden := map[string]bool{
		"cd": true, "grep": true, "ls": true, "man": true, "open": true,
	}
	for _, alias := range strings.Split(raw, ",") {
		if alias = strings.TrimSpace(alias); alias != "" {
			overridden[alias] = true
		}
	}
	return overridden
}

// FromShell expands fish functions and aliases.
func (f *Fish) FromShell(script string) string {
	aliases := f.GetAliases()
	binary := strings.SplitN(script, " ", 2)[0]
	replacement, ok := aliases[binary]
	if !ok {
		return script
	}
	if replacement != binary {
		return strings.Replace(script, binary, replacement, 1)
	}
	escaped := strings.ReplaceAll(script, dquote, backslash+dquote)
	return "fish -ic " + dquote + escaped + dquote
}

// PutToHistory appends a command to the fish history file.
func (f *Fish) PutToHistory(command string) error {
	path := homePath(fishHistoryFile())
	if _, err := os.Stat(path); err != nil {
		return nil
	}
	line := "- cmd: " + command + newline +
		"   when: " + fmt.Sprintf("%d", time.Now().Unix()) + newline
	file, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		logs.Exception("Can't update history", err)
		return nil
	}
	defer file.Close()
	if _, err := file.WriteString(line); err != nil {
		logs.Exception("Can't update history", err)
	}
	return nil
}

// HowToConfigure describes how to configure the alias for fish.
func (f *Fish) HowToConfigure() *Configuration {
	return createConfiguration(consts.AppName+" --alias | source",
		"~/.config/fish/config.fish", "fish")
}

// fishHistoryFile returns the fish history file path.
func fishHistoryFile() string {
	return "~/.config/fish/fish_history"
}

// fishParseHistoryLine extracts the command from a fish history line.
func fishParseHistoryLine(line string) string {
	if _, command, found := strings.Cut(line, "- cmd: "); found {
		return command
	}
	return ""
}

// fishVersion returns the fish version.
func fishVersion() string {
	fields := strings.Fields(runOutput("fish", "--version"))
	if len(fields) == 0 {
		return ""
	}
	return fields[len(fields)-1]
}
