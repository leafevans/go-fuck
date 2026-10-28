package rules

import (
	"bytes"
	"os/exec"
	"regexp"
	"strings"

	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// reactNativeCommandUnrecognizedPattern matches the error printed for an
// unknown command.
var reactNativeCommandUnrecognizedPattern = regexp.MustCompile(`Unrecognized command '.*'`)

// reactNativeCommandMistake matches the misspelled command.
var reactNativeCommandMistake = regexp.MustCompile(`Unrecognized command '(.*)'`)

// reactNativeCommandReadHelp runs "react-native --help" and captures its
// output.
var reactNativeCommandReadHelp = func() string {
	var stdout bytes.Buffer
	command := exec.Command("react-native", "--help")
	command.Stdout = &stdout
	_ = command.Run()
	return stdout.String()
}

// reactNativeCommandParseCommands extracts the commands from the help output.
func reactNativeCommandParseCommands(output string) []string {
	commands := []string{}
	shouldYield := false
	for _, line := range strings.Split(output, newline) {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.Contains(line, "Commands:") {
			shouldYield = true
			continue
		}
		if shouldYield {
			commands = append(commands, strings.Split(line, " ")[0])
		}
	}
	return commands
}

// reactNativeCommandCommands returns the commands of react-native.
var reactNativeCommandCommands = utils.Memoize(func() []string {
	return reactNativeCommandParseCommands(reactNativeCommandReadHelp())
})

// reactNativeCommandMatch reports whether react-native did not recognize a
// command.
func reactNativeCommandMatch(command *types.Command) bool {
	return reactNativeCommandUnrecognizedPattern.MatchString(command.Output)
}

// reactNativeCommandNewCommand suggests the closest react-native commands.
func reactNativeCommandNewCommand(command *types.Command) ([]string, error) {
	matches := reactNativeCommandMistake.FindAllStringSubmatch(command.Output, -1)
	if len(matches) == 0 {
		return nil, nil
	}
	return utils.ReplaceCommand(command, matches[0][1], reactNativeCommandCommands()), nil
}

func init() {
	register(types.NewRule("react_native_command_unrecognized",
		utils.ForApp([]string{"react-native"}, 0, reactNativeCommandMatch),
		reactNativeCommandNewCommand))
}
