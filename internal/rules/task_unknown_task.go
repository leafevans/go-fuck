package rules

import (
	"regexp"

	"github.com/leafevans/go-fuck/v2/internal/types"
	"github.com/leafevans/go-fuck/v2/internal/utils"
)

// taskUnknownTask matches the error of go-task for a missing task when it
// suggests a similar task, for example:
// task: Task "typo" does not exist. Did you mean "task"?
var taskUnknownTask = regexp.MustCompile(`task: Task "([^"]*)" does not exist\. Did you mean "([^"]*)"\?`)

// taskUnknownTaskMatch reports whether go-task failed because the task does
// not exist and suggested a similar one.
func taskUnknownTaskMatch(command *types.Command) bool {
	return taskUnknownTask.MatchString(command.Output)
}

// taskUnknownTaskNewCommand replaces the unknown task with the task that
// go-task suggests.
func taskUnknownTaskNewCommand(command *types.Command) ([]string, error) {
	match := taskUnknownTask.FindStringSubmatch(command.Output)
	if match == nil {
		return nil, nil
	}
	corrected := utils.ReplaceArgument(command.Script, match[1], match[2])
	if corrected == command.Script {
		return nil, nil
	}
	return []string{corrected}, nil
}

func init() {
	register(types.NewRule(
		"task_unknown_task",
		utils.ForApp([]string{"task", "go-task"}, 0, taskUnknownTaskMatch),
		taskUnknownTaskNewCommand,
	))
}
