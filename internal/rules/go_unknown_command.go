package rules

import (
	"bytes"
	"os/exec"
	"strings"

	"github.com/leafevans/go-fuck/v2/internal/types"
	"github.com/leafevans/go-fuck/v2/internal/utils"
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
	if index+2 >= len(lines) {
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
	parts := command.ScriptParts()
	if len(parts) < 2 || strings.HasPrefix(parts[1], "-") {
		return false
	}
	if goCommandGroup(parts[1]) && (len(parts) < 3 || strings.HasPrefix(parts[2], "-")) {
		return false
	}
	want := "go " + parts[1] + ": unknown command"
	for line := range strings.SplitSeq(command.Output, "\n") {
		if strings.TrimSpace(line) == want {
			return true
		}
	}
	return false
}

func goCommandGroup(name string) bool { return name == "mod" || name == "work" }

// goUnknownCommandNewCommand replaces the unknown command with the closest
// one.
func goUnknownCommandNewCommand(command *types.Command) ([]string, error) {
	words := simpleCommandWords(command.Script)
	if len(words) < 2 || strings.HasPrefix(words[1].text, "-") {
		return nil, nil
	}
	index := 1
	var candidates []string
	if goCommandGroup(words[1].text) {
		if len(words) < 3 || strings.HasPrefix(words[2].text, "-") {
			return nil, nil
		}
		index = 2
		candidates = goUnknownCommandParseCommands(goHelpOutput(words[1].text))
	} else {
		candidates = goUnknownCommandCommands()
	}
	target := words[index].text
	closest := utils.GetClosest(target, candidates, 0.6, false)
	if index == 2 && target == words[1].text {
		return nil, nil
	}
	return replaceCommandWord(command.Script, words[index], "", closest), nil
}

func init() {
	register(types.NewRule("go_unknown_command",
		utils.ForApp([]string{"go"}, 0, goUnknownCommandMatch),
		goUnknownCommandNewCommand))
}
