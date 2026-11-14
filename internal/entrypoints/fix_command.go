package entrypoints

import (
	"errors"
	"io"
	"os"
	"strings"

	"github.com/leafevans/go-fuck/v2/internal/argumentparser"
	"github.com/leafevans/go-fuck/v2/internal/conf"
	"github.com/leafevans/go-fuck/v2/internal/consts"
	"github.com/leafevans/go-fuck/v2/internal/corrector"
	"github.com/leafevans/go-fuck/v2/internal/logs"
	"github.com/leafevans/go-fuck/v2/internal/shells"
	"github.com/leafevans/go-fuck/v2/internal/types"
	"github.com/leafevans/go-fuck/v2/internal/ui"
	"github.com/leafevans/go-fuck/v2/internal/utils"
)

// fixCommand fixes the previous command and prints the corrected command on
// standard output.
func fixCommand(args *argumentparser.Args) int {
	settings, err := conf.Load(&conf.Args{Yes: args.Yes, Debug: args.Debug, Repeat: args.Repeat})
	if err != nil {
		logs.Exception("Can't load settings", err)
	} else {
		warnUnknownSettings(settings)
	}
	loadCustomRules()
	warnOutdatedAlias()

	stopTimer := logs.DebugTime("Total")
	defer stopTimer()
	logs.Debug("Run with settings: %+v", conf.Current())

	rawCommand := getRawCommand(args)
	if len(rawCommand) == 0 {
		logs.Debug("Empty command, nothing to do")
		if args.JSON {
			return printJSON(nil, nil)
		}
		return 0
	}

	var command *types.Command
	if args.OutputFrom != "" {
		commandOutput, readErr := readCommandOutput(args.OutputFrom)
		if readErr != nil {
			logs.Exception("Can't read the command output", readErr)
			return 1
		}
		command, err = types.FromScriptAndOutput(rawCommand, commandOutput)
	} else {
		command, err = types.FromRawScript(rawCommand)
	}
	if errors.Is(err, types.ErrEmptyCommand) {
		logs.Debug("Empty command, nothing to do")
		if args.JSON {
			return printJSON(nil, nil)
		}
		return 0
	}
	if err != nil {
		logs.Exception("Can't get output of the command", err)
		return 1
	}

	corrected := corrector.GetCorrectedCommands(command)

	if args.JSON {
		return printJSON(command, corrected)
	}

	selected := ui.SelectCommand(corrected)
	if selected == nil {
		return 1
	}
	selected.Run(command)
	return 0
}

// readCommandOutput returns the output of the failed command from a file, or
// from standard input when the file is "-".
//
// The text is normalized like captured output: a UTF-8 byte order mark is
// dropped and CRLF line endings become LF, because the rules are written
// against LF and a stray carriage return breaks expressions anchored to the
// end of a line.
func readCommandOutput(path string) (string, error) {
	var output []byte
	var err error
	if path == "-" {
		output, err = io.ReadAll(os.Stdin)
	} else {
		output, err = os.ReadFile(path)
	}
	if err != nil {
		return "", err
	}
	text := strings.TrimPrefix(string(output), "\uFEFF")
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "")
	return text, nil
}

// warnUnknownSettings reports top level settings file keys that go-fuck
// ignores, so a typo such as "rule" instead of "rules" does not go unnoticed.
func warnUnknownSettings(settings *conf.Settings) {
	if settings == nil || len(settings.UnknownKeys) == 0 {
		return
	}
	logs.Warn("Unknown settings in settings.json (ignored): " + strings.Join(settings.UnknownKeys, ", "))
}

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

	for i := len(lines) - 1; i >= 0; i-- {
		command := lines[i]
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
	lines := strings.Split(payload, "\n")
	last := -1
	for i, line := range lines {
		if strings.HasPrefix(line, "\t ") {
			last = i
		}
	}
	if last < 0 {
		return singleMarkedCommand(payload)
	}
	lastEntry := trimFCEntry(strings.Join(lines[last:], "\n"))
	if isAliasLike(lastEntry) {
		if last == 0 {
			return nil
		}
		previous := trimFCEntry(strings.Join(lines[:last], "\n"))
		if previous == "" {
			return nil
		}
		return []string{previous}
	}
	if lastEntry == "" {
		return nil
	}
	return []string{lastEntry}
}

// trimFCEntry removes the tab and space that fc adds to the first line of an
// entry, and the line breaks that command substitution keeps at the end.
func trimFCEntry(entry string) string {
	entry = strings.TrimPrefix(entry, "\t ")
	return strings.TrimRight(entry, "\r\n")
}

// isAliasLike reports whether a single line command looks like a call to the
// alias itself.
func isAliasLike(command string) bool {
	if strings.Contains(command, "\n") {
		return false
	}
	return utils.SequenceRatio(conf.Alias(), command) >= consts.DiffWithAlias
}

// zshMarkedRawCommand returns the previous entry that the zsh alias reads from
// the history array. The value is already the raw command, including line
// breaks.
func zshMarkedRawCommand(payload string) []string {
	command := strings.TrimRight(payload, "\r\n")
	if command == "" || isAliasLike(command) {
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
	if !strings.Contains(command, string(rune(10))) {
		if utils.SequenceRatio(conf.Alias(), command) >= consts.DiffWithAlias {
			return nil
		}
	}
	return []string{command}
}
