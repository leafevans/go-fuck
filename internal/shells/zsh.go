package shells

import (
	"fmt"
	"strings"

	"github.com/leafevans/go-fuck/v2/internal/conf"
	"github.com/leafevans/go-fuck/v2/internal/consts"
)

// zshAliasTemplate defines the alias function for zsh.
const zshAliasTemplate = `%[1]s () {
    export TF_SHELL=zsh;
    export TF_ALIAS=%[1]s;
    export TF_ALIAS_VERSION=2;
    TF_SHELL_ALIASES=$(alias);
    export TF_SHELL_ALIASES;
    TF_HISTORY=$'\x01'"${history[$((HISTCMD-1))]}";
    export TF_HISTORY;
    TF_CMD=$(
        %[2]s %[3]s $@
    ) && eval $TF_CMD;
    unset TF_HISTORY;
%[4]s
}`

// Zsh implements the zsh specific behavior.
type Zsh struct {
	*Generic
}

// NewZsh returns a zsh shell.
func NewZsh() *Zsh {
	generic := newGeneric("zsh", "ZSH")
	generic.historyFile = zshHistoryFile
	generic.parseHistoryLine = zshParseHistoryLine
	generic.versionFn = zshVersion
	return &Zsh{Generic: generic}
}

// AppAlias returns the zsh alias function.
func (z *Zsh) AppAlias(aliasName string) string {
	alterHistory := ""
	if conf.Current().AlterHistory {
		alterHistory = "    test -n " + dquote + "$TF_CMD" + dquote + " && print -s $TF_CMD"
	}
	return fmt.Sprintf(zshAliasTemplate, aliasName, consts.ExecutableName(),
		consts.ArgumentPlaceholder, alterHistory)
}

// InstantModeAlias returns the alias of the instant mode.
//
// In a normal shell it returns the code that starts the logged shell. Inside
// the logged shell it returns the regular alias with a prompt marker so that
// go-fuck can find the command output in the terminal log.
func (z *Zsh) InstantModeAlias(aliasName string) string {
	if instantModeEnabled() {
		return instantModeMarkedAlias(z.AppAlias(aliasName), "%{", "%}", zshInstantModeMarkHook)
	}
	return instantModeInitAlias()
}

// FromShell expands the aliases exported by the zsh alias.
func (z *Zsh) FromShell(script string) string {
	return z.expandAliases(script, z.GetAliases())
}

// GetAliases returns the aliases that were exported by the zsh alias.
func (z *Zsh) GetAliases() map[string]string {
	aliases := map[string]string{}
	for _, line := range strings.Split(env("TF_SHELL_ALIASES"), newline) {
		if line == "" || !strings.Contains(line, "=") {
			continue
		}
		if name, value, ok := parseAliasLine(line, "="); ok {
			aliases[name] = value
		}
	}
	return aliases
}

// HowToConfigure describes how to configure the alias for zsh.
func (z *Zsh) HowToConfigure() *Configuration {
	configuration := createConfiguration("eval "+dquote+"$("+consts.ExecutableName()+" --alias)"+dquote,
		"~/.zshrc", "source ~/.zshrc")
	configuration.InstantContent = "eval " + dquote + "$(" + consts.ExecutableName() +
		" --alias --enable-experimental-instant-mode)" + dquote
	return configuration
}

// zshHistoryFile returns the zsh history file path.
func zshHistoryFile() string {
	if histfile := env("HISTFILE"); histfile != "" {
		return histfile
	}
	return "~/.zsh_history"
}

// zshParseHistoryLine extracts the command from an extended zsh history line.
func zshParseHistoryLine(line string) string {
	if _, command, found := strings.Cut(line, ";"); found {
		return command
	}
	return ""
}

// zshVersion returns the zsh version.
func zshVersion() string {
	return runOutput("zsh", "-c", "echo $ZSH_VERSION")
}

// zshInstantModeMarkHook re-adds the prompt marker before every prompt.
//
// The hook is appended to precmd_functions, so it runs after the user prompt
// functions that rebuild PS1 (starship, powerlevel10k, ...).
const zshInstantModeMarkHook = `__gofuck_mark_prompt() {
    case "$PS1" in
        *"$__gofuck_mark"*) ;;
        *) PS1="%{${__gofuck_mark}%}$PS1" ;;
    esac
}
precmd_functions+=(__gofuck_mark_prompt)`
