package entrypoints

import (
	"os"
	"slices"
	"strings"

	"github.com/leafevans/go-fuck/v4/internal/argumentparser"
	"github.com/leafevans/go-fuck/v4/internal/conf"
	"github.com/leafevans/go-fuck/v4/internal/consts"
	"github.com/leafevans/go-fuck/v4/internal/logs"
	"github.com/leafevans/go-fuck/v4/internal/shells"
	"github.com/leafevans/go-fuck/v4/internal/utils"
)

// warnOutdatedAlias warns when a bash or zsh alias still passes the old
// line-based history value, which cannot restore multi-line commands.
func warnOutdatedAlias() {
	if os.Getenv("TF_HISTORY") == "" {
		return
	}
	reload := ""
	switch shells.Current().Name() {
	case "bash", "zsh":
		reload = "eval \"$(" + consts.ExecutableName() + " --alias)\""
	case "powershell":
		reload = "iex \"$(" + consts.ExecutableName() + " --alias)\""
	default:
		return
	}
	if os.Getenv(consts.AliasVersionEnv) == consts.AliasVersion {
		return
	}
	if strings.HasPrefix(os.Getenv("TF_HISTORY"), rawCommandMarker) {
		return
	}
	logs.Warn("Your shell alias is outdated; run " + reload +
		" to enable multi-line commands and the information commands (--list-rules, --json).")
}

// getRawCommand returns the command line that should be fixed.
//
// When go-fuck is invoked from a shell alias without explicit arguments, the
// command is looked up in the history exported by the alias. The most recent
// history entry that is neither a call to go-fuck nor a known executable is
// used.
func getRawCommand(args *argumentparser.Args) []string {
	if args.ForceCommand != "" {
		return []string{args.ForceCommand}
	}

	history := os.Getenv("TF_HISTORY")
	if history == "" {
		return args.Command
	}

	// The bash and zsh aliases mark the history value and pass only the
	// previous entry. A command that was typed over several lines must be
	// returned as one value, not split into lines.
	if strings.HasPrefix(history, rawCommandMarker) {
		return markedRawCommand(history)
	}

	// The PowerShell alias passes the previous entry of Get-History as a single
	// value, and PowerShell keeps a command that was typed over several lines
	// as one entry with embedded line breaks. Splitting it into lines would
	// return only the last line of a block or backtick continuation.
	if shells.Current().Name() == "powershell" {
		return powerShellRawCommand(history)
	}

	lines := strings.Split(history, string(rune(10)))
	alias := conf.Alias()

	for _, command := range slices.Backward(lines) {
		if utils.HistoryExecutable(command) == alias {
			continue
		}
		diff := utils.SequenceRatio(alias, command)
		if diff < consts.DiffWithAlias || utils.IsExecutable(command) {
			return []string{command}
		}
	}
	return nil
}

// rawCommandMarker prefixes the history value that the bash and zsh aliases
// pass. Unlike the older line based history, it marks the value as the
// previous entry, so multi-line commands can be kept as one value.
const rawCommandMarker = "\x01"

// markedRawCommand restores the previous command from the value passed by the
// bash or zsh alias.
func markedRawCommand(history string) []string {
	payload := strings.TrimPrefix(history, rawCommandMarker)
	var commands []string
	switch shells.Current().Name() {
	case "bash":
		commands = bashMarkedRawCommand(payload)
	case "zsh":
		commands = zshMarkedRawCommand(payload)
	default:
		commands = singleMarkedCommand(payload)
	}
	return sanitizeMarkedCommands(commands)
}

// sanitizeMarkedCommands removes the internal marker from parsed commands so
// that a history value containing the control character cannot leak it.
func sanitizeMarkedCommands(commands []string) []string {
	filtered := commands[:0]
	for _, command := range commands {
		command = strings.ReplaceAll(command, rawCommandMarker, "")
		if command != "" {
			filtered = append(filtered, command)
		}
	}
	return filtered
}

// bashMarkedRawCommand extracts the previous command from the output of
// "fc -ln -2". fc prefixes the first line of every entry with a tab and a
// space, while continuation lines keep their own indentation. The last entry
// is usually the alias invocation and is dropped; when the alias call is not
// in the history, the last entry is the previous command.
func bashMarkedRawCommand(payload string) []string {
	const prefix = "\t "
	for end := len(payload); ; {
		start := strings.LastIndex(payload[:end], "\n"+prefix)
		if start < 0 && !strings.HasPrefix(payload, prefix) {
			if end == len(payload) {
				return singleMarkedCommand(payload)
			}
			return nil
		}
		entry := strings.TrimRight(payload[start+1+len(prefix):end], "\r\n")
		if entry != "" && !isAliasCall(entry) {
			return []string{entry}
		}
		if start < 0 {
			return nil
		}
		end = start
	}
}

// isAliasCall excludes the alias itself, never a whole multiline command.
func isAliasCall(command string) bool {
	return !strings.Contains(command, "\n") && utils.HistoryExecutable(command) == conf.Alias()
}

// zshMarkedRawCommand returns the previous entry that the zsh alias reads from
// the history array. The value is already the raw command, including line
// breaks.
func zshMarkedRawCommand(payload string) []string {
	command := strings.TrimRight(payload, "\r\n")
	if command == "" || isAliasCall(command) {
		return nil
	}
	return []string{command}
}

// singleMarkedCommand returns a marked history value as one command.
func singleMarkedCommand(payload string) []string {
	command := strings.TrimRight(payload, "\r\n")
	if command == "" {
		return nil
	}
	return []string{command}
}

// powerShellRawCommand returns the command to fix from the history value of the
// PowerShell alias.
//
// Unlike the POSIX aliases, the PowerShell alias does not pass a slice of the
// history: it passes the previous Get-History entry. A command typed over
// several lines (a block or a backtick continuation) is a single entry that
// contains line breaks, so the value is used as it is. A trailing line break
// is trimmed because PowerShell keeps it for multi line entries.
func powerShellRawCommand(history string) []string {
	command := strings.TrimRight(history, "\r\n")
	if command == "" {
		return nil
	}
	// The previous command can be a call to the alias itself, for example when
	// the alias is invoked twice in a row.
	if !strings.Contains(command, "\n") && strings.EqualFold(utils.HistoryExecutable(command), conf.Alias()) {
		return nil
	}
	return []string{command}
}
