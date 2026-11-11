package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/internal/types"
)

// cmdMissingFileCommands are the cmd.exe commands that read or move an
// existing file.
var cmdMissingFileCommands = []string{"attrib", "copy", "move", "del", "erase", "type", "more", "ren", "rename"}

// cmdMissingFileErrors are the cmd.exe errors for a missing file.
var cmdMissingFileErrors = []string{
	"The system cannot find the file specified.",
	"Could Not Find ",
	"File not found - ",
}

// cmdMissingFileMatch reports whether a command failed because a file it names
// does not exist.
func cmdMissingFileMatch(command *types.Command) bool {
	if !cmdCommandIn(cmdFirstWord(command.Script), cmdMissingFileCommands) {
		return false
	}
	matched := false
	for _, message := range cmdMissingFileErrors {
		if strings.Contains(command.Output, message) {
			matched = true
			break
		}
	}
	if !matched {
		return false
	}
	_, argument := cmdFirstRawArgument(command.Script)
	if argument == "" || strings.ContainsAny(argument, "*?") {
		return false
	}
	return !cmdFileExists(argument)
}

// cmdMissingFileNewCommand replaces the missing file with the closest existing
// one in the same directory.
func cmdMissingFileNewCommand(command *types.Command) ([]string, error) {
	raw, argument := cmdFirstRawArgument(command.Script)
	if raw == "" || cmdFileExists(argument) {
		return nil, nil
	}
	closest := cmdClosestFile(argument)
	if closest == "" {
		return nil, nil
	}
	return []string{cmdReplaceRawArgument(command.Script, raw, closest)}, nil
}

// cmdCommandIn reports whether name is one of commands.
func cmdCommandIn(name string, commands []string) bool {
	for _, command := range commands {
		if strings.EqualFold(name, command) {
			return true
		}
	}
	return false
}

func init() {
	register(types.NewRule("cmd_missing_file",
		cmdMissingFileMatch,
		cmdMissingFileNewCommand))
}
