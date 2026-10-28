package rules

import (
	"os/exec"
	"regexp"
	"strings"

	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// yarnCommandNotFoundPattern matches the missing command error of yarn.
var yarnCommandNotFoundPattern = regexp.MustCompile(`error Command "(.*)" not found\.`)

// yarnCommandNotFoundAliases maps npm commands to their yarn equivalents.
var yarnCommandNotFoundAliases = map[string]string{"require": "add"}

// yarnCommandNotFoundParseTasks extracts the commands from "yarn --help" output.
func yarnCommandNotFoundParseTasks(commandOutput string) []string {
	tasks := []string{}
	shouldYield := false
	for _, line := range strings.Split(commandOutput, "\n") {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "Commands:") {
			shouldYield = true
			continue
		}
		if shouldYield && strings.Contains(line, "- ") {
			parts := strings.Split(line, " ")
			tasks = append(tasks, parts[len(parts)-1])
		}
	}
	return tasks
}

// yarnAllTasks returns the yarn commands reported by "yarn --help".
//
// The result is memoized because yarn is started once per go-fuck invocation.
var yarnAllTasks = utils.Memoize(func() []string {
	output, err := exec.Command("yarn", "--help").Output()
	if err != nil && len(output) == 0 {
		return nil
	}
	return yarnCommandNotFoundParseTasks(string(output))
})

// yarnCommandNotFoundMatch reports whether yarn reported an unknown command.
func yarnCommandNotFoundMatch(command *types.Command) bool {
	return yarnCommandNotFoundPattern.FindStringSubmatch(command.Output) != nil
}

// yarnCommandNotFoundNewCommand suggests the closest yarn command.
func yarnCommandNotFoundNewCommand(command *types.Command) ([]string, error) {
	match := yarnCommandNotFoundPattern.FindStringSubmatch(command.Output)
	if match == nil {
		return nil, nil
	}
	misspelledTask := match[1]
	if yarnCommand, ok := yarnCommandNotFoundAliases[misspelledTask]; ok {
		return []string{utils.ReplaceArgument(command.Script, misspelledTask, yarnCommand)}, nil
	}
	return utils.ReplaceCommand(command, misspelledTask, yarnAllTasks()), nil
}

func init() {
	register(types.NewRule("yarn_command_not_found",
		utils.ForApp([]string{"yarn"}, 0, yarnCommandNotFoundMatch),
		yarnCommandNotFoundNewCommand))
}
