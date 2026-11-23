package shells

import (
	"fmt"
	"strings"
	"sync"

	"github.com/leafevans/go-fuck/v3/internal/consts"
	"github.com/leafevans/go-fuck/v3/internal/logs"
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
//
// The failed command and the corrected command are captured in double quotes:
// an unquoted command substitution word splits on every blank and expands
// globs, so "make *.txt" would reach go-fuck as the files of the directory and
// "echo  a" would lose the second blank. The raw command is handed over in
// TF_HISTORY so go-fuck sees exactly what was typed, and the variable is
// cleared again so a later manual invocation does not reuse a stale history.
func (t *Tcsh) AppAlias(aliasName string) string {
	return fmt.Sprintf(
		"alias %[1]s 'setenv TF_SHELL tcsh && setenv TF_ALIAS %[1]s && setenv TF_ALIAS_VERSION %[4]s && "+
			"set fucked_cmd = \"%[3]shistory -h 2 | head -n 1%[3]s\" && "+
			"setenv TF_HISTORY \"$fucked_cmd\" && "+
			"set unfucked_cmd = \"%[3]s%[2]s%[3]s\" && "+
			"eval \"$unfucked_cmd\" ; unsetenv TF_HISTORY'",
		aliasName, consts.ExecutableName(), backtick, consts.AliasVersion)
}

// Quote returns a tcsh escaped version of s.
//
// csh performs history substitution even inside single quotes, so an
// exclamation mark has to be escaped, and a plain newline ends the quoted
// string: both are written with a leading backslash. An embedded single quote
// closes the string, escapes the quote and opens it again.
func (t *Tcsh) Quote(s string) string {
	var builder strings.Builder
	builder.WriteString("'")
	for _, character := range s {
		switch character {
		case '!':
			builder.WriteString("\\!")
		case '\n':
			builder.WriteString("\\\n")
		case '\'':
			builder.WriteString("'\\''")
		default:
			builder.WriteRune(character)
		}
	}
	builder.WriteString("'")
	return builder.String()
}

// SplitCommand splits a tcsh command line. Tcsh keeps history expansion active
// inside single quotes, so Quote escapes "!" and newlines with a backslash;
// those escapes have to be decoded here.
func (t *Tcsh) SplitCommand(command string) []string {
	return splitTcsh(command)
}

// splitTcsh splits command into arguments following tcsh quoting and escaping rules.
func splitTcsh(command string) []string {
	var (
		args    []string
		current strings.Builder
		inWord  bool
	)
	flush := func() {
		if inWord {
			args = append(args, current.String())
			current.Reset()
			inWord = false
		}
	}
	runes := []rune(command)
	for index := 0; index < len(runes); index++ {
		switch ch := runes[index]; {
		case ch == ' ' || ch == '\t' || ch == '\r' || ch == '\n':
			flush()
		case ch == '\'':
			inWord = true
			for index++; index < len(runes); index++ {
				if runes[index] == '\'' {
					break
				}
				if runes[index] == '\\' && index+1 < len(runes) &&
					(runes[index+1] == '!' || runes[index+1] == '\n') {
					index++
				}
				current.WriteRune(runes[index])
			}
		case ch == '"':
			inWord = true
			for index++; index < len(runes); index++ {
				if runes[index] == '"' {
					break
				}
				if runes[index] == '\\' && index+1 < len(runes) {
					index++
				}
				current.WriteRune(runes[index])
			}
		case ch == '\\' && index+1 < len(runes):
			index++
			current.WriteRune(runes[index])
			inWord = true
		default:
			current.WriteRune(ch)
			inWord = true
		}
	}
	flush()
	return args
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
		for line := range strings.SplitSeq(runOutput("tcsh", "-ic", "alias"), newline) {
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
	return createConfiguration("eval "+backtick+consts.ExecutableName()+" --alias"+backtick,
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
