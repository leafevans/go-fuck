package rules

import (
	"os/exec"
	"strings"

	"github.com/leafevans/go-fuck/v3/internal/specific"
	"github.com/leafevans/go-fuck/v3/internal/types"
	"github.com/leafevans/go-fuck/v3/internal/utils"
)

// npmWrongCommandListing is the line that starts the npm command listing in the
// npm 6 and older output.
const npmWrongCommandListing = "where <command> is one of:"

// npmWrongCommandUnknown is the line that starts the npm 7 and newer output.
const npmWrongCommandUnknown = "Unknown command: "

// npmWrongCommandHelpCommands returns the commands listed by "npm help". It is
// used for npm versions that no longer print the command list in the error
// output.
var npmWrongCommandHelpCommands = utils.Memoize(func() []string {
	output, err := exec.Command("npm", "help").Output()
	if err != nil && len(output) == 0 {
		return nil
	}
	return npmWrongCommandParseHelp(string(output))
})

func npmGlobalOption(name string) commandOption {
	switch name {
	case "--prefix", "--cache", "--registry", "--userconfig", "--globalconfig", "--loglevel", "--workspace", "-w":
		return valueOption
	case "--global", "-g", "--force", "-f", "--silent", "-s", "--verbose", "--workspaces", "--no-workspaces":
		return flagOption
	default:
		return unknownOption
	}
}

// npmWrongCommandMatch reports whether an unknown npm command was used.
func npmWrongCommandMatch(command *types.Command) bool {
	words := simpleCommandWords(command.Script)
	if len(words) == 0 || words[0].text != "npm" || subcommandIndex(words, 1, npmGlobalOption) < 0 {
		return false
	}
	return strings.Contains(command.Output, npmWrongCommandListing) ||
		strings.Contains(command.Output, npmWrongCommandUnknown)
}

// npmWrongCommandAvailable returns the npm commands listed in the help output.
func npmWrongCommandAvailable(commandOutput string) []string {
	commands := []string{}
	listing := false
	for line := range strings.SplitSeq(commandOutput, "\n") {
		if strings.HasPrefix(line, npmWrongCommandListing) {
			listing = true
			continue
		}
		if !listing {
			continue
		}
		if line == "" {
			break
		}
		for command := range strings.SplitSeq(line, ", ") {
			if trimmed := strings.TrimSpace(command); trimmed != "" {
				commands = append(commands, trimmed)
			}
		}
	}
	return commands
}

// npmWrongCommandParseHelp parses the command list of the "npm help" output
// that npm 7 and newer print.
func npmWrongCommandParseHelp(output string) []string {
	commands := []string{}
	listing := false
	for line := range strings.SplitSeq(output, newline) {
		trimmed := strings.TrimSpace(line)
		if trimmed == "All commands:" {
			listing = true
			continue
		}
		if !listing {
			continue
		}
		if trimmed == "" {
			// "npm help" prints a blank line between the header and the list.
			if len(commands) > 0 {
				break
			}
			continue
		}
		for command := range strings.SplitSeq(trimmed, ",") {
			if candidate := strings.TrimSpace(command); candidate != "" {
				commands = append(commands, candidate)
			}
		}
	}
	return commands
}

// npmWrongCommandAvailableCommands returns the commands that the suggestion is
// chosen from. npm 6 and older embed the list in the error output; npm 7 and
// newer only point at "npm help", so the help output is used as a fallback.
func npmWrongCommandAvailableCommands(commandOutput string) []string {
	if commands := npmWrongCommandAvailable(commandOutput); len(commands) > 0 {
		return commands
	}
	return npmWrongCommandHelpCommands()
}

// npmWrongCommandNewCommand replaces the unknown command with the closest npm
// command.
func npmWrongCommandNewCommand(command *types.Command) ([]string, error) {
	words := simpleCommandWords(command.Script)
	index := subcommandIndex(words, 1, npmGlobalOption)
	if index < 0 {
		return nil, nil
	}
	wrongCommand := words[index].text
	for line := range strings.SplitSeq(command.Output, "\n") {
		if reported, ok := strings.CutPrefix(strings.TrimSpace(line), npmWrongCommandUnknown); ok &&
			strings.Trim(reported, `"`) != wrongCommand {
			return nil, nil
		}
	}
	fixed := utils.GetClosestDefault(wrongCommand, npmWrongCommandAvailableCommands(command.Output))
	return replaceCommandWord(command.Script, words[index], "", fixed), nil
}

func init() {
	// Python: enabled_by_default = npm_available
	rule := types.NewRule("npm_wrong_command",
		specific.SudoSupportMatch(utils.ForApp([]string{"npm"}, 0, npmWrongCommandMatch)),
		specific.SudoSupportNew(npmWrongCommandNewCommand))
	rule.DefaultProbe = startupTools.NpmAvailable
	register(rule)
}
