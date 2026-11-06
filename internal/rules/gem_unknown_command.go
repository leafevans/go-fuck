package rules

import (
	"os/exec"
	"regexp"
	"strings"

	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// gemUnknownCommandRegexp matches the unknown command name.
var gemUnknownCommandRegexp = regexp.MustCompile(`Unknown command ([^\n]*)`)

// gemUnknownCommandGemCommands returns the commands reported by
// "gem help commands".
var gemUnknownCommandGemCommands = utils.Memoize(func() []string {
	output, err := exec.Command("gem", "help", "commands").Output()
	if err != nil {
		return nil
	}
	return gemUnknownCommandParseCommands(string(output))
})

// gemUnknownCommandParseCommands parses the command list of the gem help
// output.
func gemUnknownCommandParseCommands(output string) []string {
	commands := []string{}
	for _, line := range strings.Split(output, "\n") {
		if !strings.HasPrefix(line, "    ") {
			continue
		}
		if fields := strings.Fields(line); len(fields) > 0 {
			commands = append(commands, fields[0])
		}
	}
	return commands
}

// gemUnknownCommandMatch reports whether gem did not recognize the command.
//
// RubyGems 3 and older raise Gem::CommandLineError; RubyGems 4 raises
// Gem::UnknownCommandError instead.
func gemUnknownCommandMatch(command *types.Command) bool {
	output := command.Output
	if !strings.Contains(output, "Unknown command") {
		return false
	}
	return strings.Contains(output, "Gem::CommandLineError") ||
		strings.Contains(output, "Gem::UnknownCommandError")
}

// gemUnknownCommandNewCommand suggests the closest valid gem commands.
func gemUnknownCommandNewCommand(command *types.Command) ([]string, error) {
	match := gemUnknownCommandRegexp.FindStringSubmatch(command.Output)
	if match == nil {
		return nil, nil
	}
	// Windows output ends the line with CRLF; the capture must not keep the CR.
	return utils.ReplaceCommand(command, strings.TrimSpace(match[1]), gemUnknownCommandGemCommands()), nil
}

func init() {
	register(types.NewRule("gem_unknown_command",
		utils.ForApp([]string{"gem"}, 0, gemUnknownCommandMatch),
		gemUnknownCommandNewCommand))
}
