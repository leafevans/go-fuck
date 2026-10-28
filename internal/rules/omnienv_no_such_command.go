package rules

import (
	"os/exec"
	"regexp"
	"strings"

	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// omnienvNoSuchCommandSupportedApps lists the version managers supported by
// the rule.
var omnienvNoSuchCommandSupportedApps = []string{"goenv", "nodenv", "pyenv", "rbenv"}

// omnienvNoSuchCommandCommonTypos maps a broken command to the replacements
// that are always suggested.
var omnienvNoSuchCommandCommonTypos = map[string][]string{
	"list":   {"versions", "install --list"},
	"remove": {"uninstall"},
}

// omnienvNoSuchCommandRegexp matches the quoted missing command.
var omnienvNoSuchCommandRegexp = regexp.MustCompile("env: no such command ['`]([^']*)'")

// omnienvNoSuchCommandAppCommands returns the commands supported by a version
// manager.
var omnienvNoSuchCommandAppCommands = utils.Memoize1(func(app string) []string {
	output, err := exec.Command(app, "commands").Output()
	if err != nil {
		return nil
	}
	commands := []string{}
	for _, line := range strings.Split(string(output), "\n") {
		if command := strings.TrimSpace(line); command != "" {
			commands = append(commands, command)
		}
	}
	return commands
})

// omnienvNoSuchCommandMatch reports whether a version manager rejected an
// unknown command.
func omnienvNoSuchCommandMatch(command *types.Command) bool {
	return strings.Contains(command.Output, "env: no such command ")
}

// omnienvNoSuchCommandNewCommand suggests the closest valid commands.
func omnienvNoSuchCommandNewCommand(command *types.Command) ([]string, error) {
	match := omnienvNoSuchCommandRegexp.FindStringSubmatch(command.Output)
	if match == nil {
		return nil, nil
	}
	broken := match[1]
	results := []string{}
	for _, typo := range omnienvNoSuchCommandCommonTypos[broken] {
		results = append(results, utils.ReplaceArgument(command.Script, broken, typo))
	}
	parts := command.ScriptParts()
	if len(parts) > 0 {
		commands := omnienvNoSuchCommandAppCommands(parts[0])
		results = append(results, utils.ReplaceCommand(command, broken, commands)...)
	}
	return results, nil
}

func init() {
	enabled := false
	for _, app := range omnienvNoSuchCommandSupportedApps {
		if utils.Which(app) != "" {
			enabled = true
			break
		}
	}
	rule := types.NewRule("omnienv_no_such_command",
		utils.ForApp(omnienvNoSuchCommandSupportedApps, 1, omnienvNoSuchCommandMatch),
		omnienvNoSuchCommandNewCommand)
	rule.EnabledByDefault = enabled
	register(rule)
}
