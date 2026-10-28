package rules

import (
	"os/exec"
	"regexp"
	"strings"

	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// gruntTaskNotFoundPattern matches the unknown grunt task reported by grunt.
var gruntTaskNotFoundPattern = regexp.MustCompile(`Warning: Task "(.*)" not found.`)

// gruntTaskNotFoundParseTasks extracts the tasks from "grunt --help" output.
func gruntTaskNotFoundParseTasks(helpOutput string) []string {
	tasks := []string{}
	shouldYield := false
	for _, line := range strings.Split(helpOutput, "\n") {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "Available tasks") {
			shouldYield = true
			continue
		}
		if shouldYield && line == "" {
			return tasks
		}
		if strings.Contains(line, "  ") {
			tasks = append(tasks, strings.Split(line, " ")[0])
		}
	}
	return tasks
}

// gruntTasks returns the tasks reported by "grunt --help".
//
// The result is memoized because grunt is started once per go-fuck invocation.
var gruntTasks = utils.Memoize(func() []string {
	output, err := exec.Command("grunt", "--help").Output()
	if err != nil && len(output) == 0 {
		return nil
	}
	return gruntTaskNotFoundParseTasks(string(output))
})

// gruntTaskNotFoundMatch reports whether grunt reported an unknown task.
func gruntTaskNotFoundMatch(command *types.Command) bool {
	return gruntTaskNotFoundPattern.FindStringSubmatch(command.Output) != nil
}

// gruntTaskNotFoundNewCommand replaces the unknown task with the closest task.
func gruntTaskNotFoundNewCommand(command *types.Command) ([]string, error) {
	match := gruntTaskNotFoundPattern.FindStringSubmatch(command.Output)
	if match == nil {
		return nil, nil
	}
	misspelledTask := strings.Split(match[1], ":")[0]
	fixed := utils.GetClosestDefault(misspelledTask, gruntTasks())
	if fixed == "" {
		return nil, nil
	}
	return []string{strings.ReplaceAll(command.Script,
		" "+misspelledTask, " "+fixed)}, nil
}

func init() {
	register(types.NewRule("grunt_task_not_found",
		utils.ForApp([]string{"grunt"}, 0, gruntTaskNotFoundMatch),
		gruntTaskNotFoundNewCommand))
}
