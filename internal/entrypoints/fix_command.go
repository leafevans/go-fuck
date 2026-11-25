package entrypoints

import (
	"errors"
	"io"
	"os"
	"slices"
	"strings"

	"github.com/leafevans/go-fuck/v4/internal/argumentparser"
	"github.com/leafevans/go-fuck/v4/internal/conf"
	"github.com/leafevans/go-fuck/v4/internal/consts"
	"github.com/leafevans/go-fuck/v4/internal/corrector"
	"github.com/leafevans/go-fuck/v4/internal/logs"
	"github.com/leafevans/go-fuck/v4/internal/types"
	"github.com/leafevans/go-fuck/v4/internal/ui"
)

// fixCommand fixes the previous command and prints the corrected command on
// standard output.
func fixCommand(args *argumentparser.Args) int {
	settings, err := conf.Load(&conf.Args{
		Yes: args.Yes, Debug: args.Debug, Repeat: args.Repeat,
		CaptureMode: args.CaptureMode,
	})
	if err != nil {
		logs.Exception("Can't load settings", err)
	} else {
		warnUnknownSettings(settings)
		warnInvalidSettings(settings)
		warnUnknownRuleSettings(settings)
		warnConflictingRuleSettings(settings)
	}
	loadCustomRules()
	warnOutdatedAlias()

	stopTimer := logs.DebugTime("Total")
	defer stopTimer()
	logs.Debug("Run with settings: %+v", conf.Current())

	rawCommand := getRawCommand(args)
	if !slices.ContainsFunc(rawCommand, func(part string) bool { return strings.TrimSpace(part) != "" }) {
		return reportMissingCommand(args.JSON)
	}
	if hint := entryPointHint(rawCommand); hint != "" && !args.JSON {
		// The words are not corrections: report the flag instead of running
		// the word as a command, which would fail without an explanation.
		logs.Failed(hint)
		return 1
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
		command, err = types.FromRawScript(rawCommand,
			!args.NoRerun && !args.Preview && !args.Explain)
	}
	if errors.Is(err, types.ErrEmptyCommand) {
		return reportMissingCommand(args.JSON)
	}
	if err != nil {
		logs.Exception("Can't get output of the command", err)
		return 1
	}

	return handleCorrections(args, command, corrector.GetCorrectedCommands(command))
}

// handleCorrections presents or applies suggestions without fetching output again.
func handleCorrections(args *argumentparser.Args, command *types.Command, corrected []*types.CorrectedCommand) int {
	switch {
	case args.JSON:
		return printJSON(command, corrected)
	case args.Explain:
		ui.ExplainCommands(command, corrected)
		if len(corrected) == 0 {
			return 1
		}
		return 0
	case len(corrected) == 0:
		if command.HasOutput() {
			logs.Failed("No matching correction rule found.")
		} else {
			logs.Failed("No suggestion without the original command output. Provide it with --force-command and --output-from.")
		}
		return 1
	case args.Preview:
		ui.PreviewCommands(corrected)
		return 0
	}
	selected := ui.SelectCommand(corrected)
	if selected == nil {
		return 1
	}
	selected.Run(command)
	return 0
}

// entryPointHint explains a word that names an information command but was
// passed as the command to fix, such as "go-fuck doctor" instead of
// "go-fuck --doctor".
//
// The shell alias passes its arguments as the command to fix, so "fuck doctor"
// stays a correction. Only a direct call outside the alias is answered with the
// flag, because the words never name a shell command that go-fuck should run.
func entryPointHint(rawCommand []string) string {
	if os.Getenv(consts.AliasEnv) != "" || len(rawCommand) != 1 {
		return ""
	}
	name := consts.ExecutableName()
	switch strings.TrimSpace(rawCommand[0]) {
	case "setup":
		return name + " setup is not a shell command to fix: run " + name + " --setup to add the alias to your shell configuration."
	case "doctor":
		return name + " doctor is not a shell command to fix: run " + name + " --doctor to print the shell integration report."
	}
	return ""
}

// reportMissingCommand writes the empty-command result in JSON or text mode.
func reportMissingCommand(json bool) int {
	if json {
		return printJSON(nil, nil)
	}
	logs.Failed("No usable command in shell history. Specify one with --force-command.")
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

// warnInvalidSettings reports settings values that were out of range and
// replaced with the built-in defaults.
func warnInvalidSettings(settings *conf.Settings) {
	if settings == nil || len(settings.InvalidValues) == 0 {
		return
	}
	logs.Warn("Invalid settings values (using the defaults): " + strings.Join(settings.InvalidValues, ", "))
}
