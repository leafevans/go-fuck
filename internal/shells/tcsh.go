package shells

import (
	"fmt"
	"strings"
	"sync"

	"github.com/leafevans/go-fuck/internal/consts"
	"github.com/leafevans/go-fuck/internal/logs"
)

// Tcsh implements the tcsh specific behavior.
type Tcsh struct {
	*Generic

	aliasesOnce sync.Once
	aliases     map[string]string
}

// NewTcsh returns a tcsh shell.
func NewTcsh() *Tcsh {
	generic := newGeneric("tcsh", "Tcsh")
	generic.historyFile = tcshHistoryFile
	generic.parseHistoryLine = tcshParseHistoryLine
	generic.versionFn = tcshVersion
	return &Tcsh{Generic: generic}
}

// AppAlias returns the tcsh alias.
func (t *Tcsh) AppAlias(aliasName string) string {
	return fmt.Sprintf(
		"alias %[1]s 'setenv TF_SHELL tcsh && setenv TF_ALIAS %[1]s && "+
			"set fucked_cmd=%[3]shistory -h 2 | head -n 1%[3]s && "+
			"eval %[3]s%[2]s $fucked_cmd%[3]s'",
		aliasName, consts.AppName, backtick)
}

// InstantModeAlias falls back to the regular alias because instant mode is not
// supported.
func (t *Tcsh) InstantModeAlias(aliasName string) string {
	logs.Warn("Instant mode not supported by go-fuck")
	return t.AppAlias(aliasName)
}

// FromShell expands the tcsh aliases.
func (t *Tcsh) FromShell(script string) string {
	return t.expandAliases(script, t.GetAliases())
}

// GetAliases returns the tcsh aliases.
func (t *Tcsh) GetAliases() map[string]string {
	t.aliasesOnce.Do(func() {
		t.aliases = map[string]string{}
		for _, line := range strings.Split(runOutput("tcsh", "-ic", "alias"), newline) {
			line = strings.TrimSpace(line)
			if line == "" || !strings.Contains(line, string(rune(9))) {
				continue
			}
			if name, value, ok := parseAliasLine(line, string(rune(9))); ok {
				t.aliases[name] = value
			}
		}
	})
	return t.aliases
}

// HowToConfigure describes how to configure the alias for tcsh.
func (t *Tcsh) HowToConfigure() *Configuration {
	return createConfiguration("eval "+backtick+consts.AppName+" --alias"+backtick,
		"~/.tcshrc", "tcsh")
}

// tcshHistoryFile returns the tcsh history file path.
func tcshHistoryFile() string {
	if histfile := env("HISTFILE"); histfile != "" {
		return histfile
	}
	return "~/.history"
}

// tcshParseHistoryLine extracts the command from a tcsh history line.
//
// Tcsh stores the timestamp of each command in a separate "#+" prefixed line,
// which is skipped here.
func tcshParseHistoryLine(line string) string {
	if strings.HasPrefix(line, "#+") {
		return ""
	}
	return line
}

// tcshVersion returns the tcsh version.
func tcshVersion() string {
	fields := strings.Fields(runOutput("tcsh", "--version"))
	if len(fields) < 2 {
		return ""
	}
	return fields[1]
}
