package rules

import (
	"os/exec"
	"regexp"
	"strings"

	"github.com/leafevans/go-fuck/v2/internal/specific"
	"github.com/leafevans/go-fuck/v2/internal/types"
	"github.com/leafevans/go-fuck/v2/internal/utils"
)

// chocoUnknownCommandFallbackCommands lists the commands of Chocolatey 2.7.3.
// It is used when the local Chocolatey installation cannot be inspected.
var chocoUnknownCommandFallbackCommands = []string{
	"apikey", "cache", "config", "export", "feature", "find", "help", "info",
	"install", "license", "list", "new", "outdated", "pack", "pin", "push",
	"rule", "search", "setapikey", "source", "support", "template",
	"uninstall", "unpackself", "upgrade",
}

// chocoUnknownCommandRegexp matches the command that Chocolatey did not find,
// for example: Could not find a command registered that meets 'instal'.
var chocoUnknownCommandRegexp = regexp.MustCompile("Could not find a command registered that meets '([^']*)'")

// chocoUnknownCommandNameRegexp matches a command name of the help output.
var chocoUnknownCommandNameRegexp = regexp.MustCompile("^[a-z][a-z0-9_-]*$")

// chocoUnknownCommandParseCommands parses the command names of the
// "choco --help" output.
//
// Every command is listed on its own line between the name and the
// description, for example:
//
//   - install - installs packages using configured sources
//   - list - lists local packages
func chocoUnknownCommandParseCommands(output string) []string {
	commands := []string{}
	seen := make(map[string]bool)
	for _, line := range strings.Split(output, newline) {
		rest, found := strings.CutPrefix(line, " * ")
		if !found {
			continue
		}
		name, _, found := strings.Cut(rest, " - ")
		if !found {
			continue
		}
		name = strings.TrimSpace(name)
		if !chocoUnknownCommandNameRegexp.MatchString(name) || seen[name] {
			continue
		}
		seen[name] = true
		commands = append(commands, name)
	}
	return commands
}

// chocoUnknownCommandHelp returns the commands reported by "choco --help".
var chocoUnknownCommandHelp = utils.Memoize(func() []string {
	output, err := exec.Command("choco", "--help").Output()
	if err != nil {
		return append([]string{}, chocoUnknownCommandFallbackCommands...)
	}
	commands := chocoUnknownCommandParseCommands(string(output))
	if len(commands) == 0 {
		return append([]string{}, chocoUnknownCommandFallbackCommands...)
	}
	return commands
})

// chocoUnknownCommandMatch reports whether Chocolatey did not recognize the
// command.
func chocoUnknownCommandMatch(command *types.Command) bool {
	return strings.Contains(command.Output, "Could not find a command registered that meets '") &&
		chocoUnknownCommandRegexp.MatchString(command.Output)
}

// chocoUnknownCommandNewCommand suggests the closest valid Chocolatey
// commands.
func chocoUnknownCommandNewCommand(command *types.Command) ([]string, error) {
	match := chocoUnknownCommandRegexp.FindStringSubmatch(command.Output)
	if match == nil {
		return nil, nil
	}
	closest := utils.GetCloseMatches(match[1], chocoUnknownCommandHelp(), 0, 0.6)
	return utils.ReplaceArguments(command.Script, match[1], closest), nil
}

func init() {
	rule := types.NewRule(
		"choco_unknown_command",
		utils.ForApp([]string{"choco"}, 0, chocoUnknownCommandMatch),
		chocoUnknownCommandNewCommand,
	)
	rule.EnabledByDefault = specific.ChocoAvailable()
	register(rule)
}
