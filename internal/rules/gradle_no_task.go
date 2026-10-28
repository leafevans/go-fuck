package rules

import (
	"os/exec"
	"regexp"
	"strings"

	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// gradleNoTaskPattern matches the gradle error for an unknown task name.
var gradleNoTaskPattern = regexp.MustCompile("Task '(.*)' (is ambiguous|not found)")

// gradleNoTaskParseTasks extracts the task names from "gradle tasks" output.
func gradleNoTaskParseTasks(tasksOutput string) []string {
	tasks := []string{}
	shouldYield := false
	for _, line := range strings.Split(tasksOutput, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "----") {
			shouldYield = true
			continue
		}
		if line == "" {
			shouldYield = false
			continue
		}
		if shouldYield && !strings.HasPrefix(line, "All tasks runnable from root project") {
			tasks = append(tasks, strings.Split(line, " ")[0])
		}
	}
	return tasks
}

// gradleTasks returns the tasks reported by the given gradle executable.
//
// The result is memoized per executable because gradle is slow and is started
// once per go-fuck invocation.
var gradleTasks = utils.Memoize1(func(gradle string) []string {
	output, err := exec.Command(gradle, "tasks").Output()
	if err != nil && len(output) == 0 {
		return nil
	}
	return gradleNoTaskParseTasks(string(output))
})

// gradleNoTaskMatch reports whether gradle reported an unknown task.
func gradleNoTaskMatch(command *types.Command) bool {
	return gradleNoTaskPattern.FindStringSubmatch(command.Output) != nil
}

// gradleNoTaskNewCommand replaces the unknown task with the closest task.
func gradleNoTaskNewCommand(command *types.Command) ([]string, error) {
	matches := gradleNoTaskPattern.FindAllStringSubmatch(command.Output, -1)
	if len(matches) == 0 {
		return nil, nil
	}
	parts := command.ScriptParts()
	if len(parts) == 0 {
		return nil, nil
	}
	return utils.ReplaceCommand(command, matches[0][1], gradleTasks(parts[0])), nil
}

func init() {
	register(types.NewRule("gradle_no_task",
		utils.ForApp([]string{"gradle", "gradlew"}, 0, gradleNoTaskMatch),
		gradleNoTaskNewCommand))
}
