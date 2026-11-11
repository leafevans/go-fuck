package entrypoints

import (
	"errors"
	"os"
	"strings"

	"github.com/leafevans/go-fuck/internal/argumentparser"
	"github.com/leafevans/go-fuck/internal/conf"
	"github.com/leafevans/go-fuck/internal/consts"
	"github.com/leafevans/go-fuck/internal/corrector"
	"github.com/leafevans/go-fuck/internal/logs"
	"github.com/leafevans/go-fuck/internal/shells"
	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/ui"
	"github.com/leafevans/go-fuck/internal/utils"
)

// fixCommand fixes the previous command and prints the corrected command on
// standard output.
func fixCommand(args *argumentparser.Args) int {
	if _, err := conf.Load(&conf.Args{Yes: args.Yes, Debug: args.Debug, Repeat: args.Repeat}); err != nil {
		logs.Exception("Can't load settings", err)
	}
	loadCustomRules()

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

	command, err := types.FromRawScript(rawCommand)
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
