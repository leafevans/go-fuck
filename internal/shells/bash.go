package shells

import (
	"fmt"
	"strings"

	"github.com/leafevans/go-fuck/internal/conf"
	"github.com/leafevans/go-fuck/internal/consts"
	"github.com/leafevans/go-fuck/internal/logs"
)

// bashAliasTemplate defines the alias function for bash.
const bashAliasTemplate = `function %[1]s () {
    export TF_SHELL=bash;
    export TF_ALIAS=%[1]s;
    export TF_SHELL_ALIASES=$(alias);
    export TF_HISTORY=$(fc -ln -10);
    TF_CMD=$(
        %[2]s %[3]s "$@"
    ) && eval "$TF_CMD";
    unset TF_HISTORY;
%[4]s
}`

// Bash implements the bash specific behavior.
type Bash struct {
	*Generic
}

// NewBash returns a bash shell.
func NewBash() *Bash {
	generic := newGeneric("bash", "Bash")
	generic.historyFile = bashHistoryFile
	generic.versionFn = bashVersion
	return &Bash{Generic: generic}
}

// AppAlias returns the bash alias function.
func (b *Bash) AppAlias(aliasName string) string {
	alterHistory := ""
	if conf.Current().AlterHistory {
		alterHistory = "    history -s $TF_CMD;"
	}
	return fmt.Sprintf(bashAliasTemplate, aliasName, consts.AppName,
		consts.ArgumentPlaceholder, alterHistory)
}

// InstantModeAlias falls back to the regular alias because instant mode is not
// supported.
func (b *Bash) InstantModeAlias(aliasName string) string {
	logs.Warn("Instant mode not supported by go-fuck")
	return b.AppAlias(aliasName)
}

// FromShell expands the aliases exported by the bash alias.
func (b *Bash) FromShell(script string) string {
	return b.expandAliases(script, b.GetAliases())
}

// GetAliases returns the aliases that were exported by the bash alias.
func (b *Bash) GetAliases() map[string]string {
	aliases := map[string]string{}
	for _, line := range strings.Split(env("TF_SHELL_ALIASES"), newline) {
		if line == "" || !strings.Contains(line, "=") {
			continue
		}
		line = strings.TrimPrefix(line, "alias ")
		if name, value, ok := parseAliasLine(line, "="); ok {
			aliases[name] = value
		}
	}
	return aliases
}

// HowToConfigure describes how to configure the alias for bash.
func (b *Bash) HowToConfigure() *Configuration {
	config := "bash config"
	if _, err := statHome("~/.bashrc"); err == nil {
		config = "~/.bashrc"
	} else if _, err := statHome("~/.bash_profile"); err == nil {
		config = "~/.bash_profile"
	}
	return createConfiguration("eval "+dquote+"$("+consts.AppName+" --alias)"+dquote,
		config, "source "+config)
}

// bashHistoryFile returns the bash history file path.
func bashHistoryFile() string {
	if histfile := env("HISTFILE"); histfile != "" {
		return histfile
	}
	return "~/.bash_history"
}

// bashVersion returns the bash version.
func bashVersion() string {
	return runOutput("bash", "-c", "echo $BASH_VERSION")
}
