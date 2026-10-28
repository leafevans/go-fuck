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

	stopTimer := logs.DebugTime("Total")
	defer stopTimer()
	logs.Debug("Run with settings: %+v", conf.Current())

	rawCommand := getRawCommand(args)
	if len(rawCommand) == 0 {
		logs.Debug("Empty command, nothing to do")
		return 0
	}

	command, err := types.FromRawScript(rawCommand)
	if errors.Is(err, types.ErrEmptyCommand) {
		logs.Debug("Empty command, nothing to do")
		return 0
	}
	if err != nil {
		logs.Exception("Can't get output of the command", err)
		return 1
	}

	selected := ui.SelectCommand(corrector.GetCorrectedCommands(command))
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

	lines := strings.Split(history, string(rune(10)))
	alias := conf.Alias()
	executables := map[string]bool{}
	for _, executable := range utils.GetAllExecutables() {
		executables[executable] = true
	}

	for i := len(lines) - 1; i >= 0; i-- {
		command := lines[i]
		diff := utils.SequenceRatio(alias, command)
		if diff < consts.DiffWithAlias || executables[command] {
			return []string{command}
		}
	}
	return nil
}
