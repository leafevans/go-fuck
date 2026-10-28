package rules

import (
	"bytes"
	"os/exec"
	"strings"

	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// goUnknownCommandReadHelp runs "go" and captures its help output. The help is
// written to stderr.
var goUnknownCommandReadHelp = func() string {
	var stderr bytes.Buffer
	goCommand := exec.Command("go")
	goCommand.Stderr = &stderr
	_ = goCommand.Run()
	return stderr.String()
}

// goUnknownCommandParseCommands extracts the commands from the "go" help.
func goUnknownCommandParseCommands(output string) []string {
	lines := strings.Split(output, newline)
	for index := range lines {
		lines[index] = strings.TrimSpace(lines[index])
	}

	index := 0
	for index < len(lines) && lines[index] != "The commands are:" {
		index++
	}
	if index == len(lines) {
		return nil
	}

	commands := []string{}
	// Skip the header and the blank line that follows it.
	for _, line := range lines[index+2:] {
		if line == "" {
			break
		}
		commands = append(commands, strings.Split(line, " ")[0])
	}
	return commands
}

// goUnknownCommandCommands returns the commands listed by "go help".
var goUnknownCommandCommands = utils.Memoize(func() []string {
	return goUnknownCommandParseCommands(goUnknownCommandReadHelp())
})

// goUnknownCommandMatch reports whether go did not recognize a command.
func goUnknownCommandMatch(command *types.Command) bool {
	return strings.Contains(command.Output, "unknown command")
}

// goUnknownCommandNewCommand replaces the unknown command with the closest
// one.
func goUnknownCommandNewCommand(command *types.Command) ([]string, error) {
	parts := command.ScriptParts()
	if len(parts) < 2 {
		return nil, nil
	}
	closest := utils.GetClosestDefault(parts[1], goUnknownCommandCommands())
	return []string{utils.ReplaceArgument(command.Script, parts[1], closest)}, nil
}

func init() {
	register(types.NewRule("go_unknown_command",
		utils.ForApp([]string{"go"}, 0, goUnknownCommandMatch),
		goUnknownCommandNewCommand))
}
