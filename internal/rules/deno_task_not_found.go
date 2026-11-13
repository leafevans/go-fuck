package rules

import (
	"regexp"
	"strings"

	"github.com/leafevans/go-fuck/v2/internal/types"
	"github.com/leafevans/go-fuck/v2/internal/utils"
)

var (
	// denoTaskNotFound matches the error line of deno for a task that does not
	// exist, for example: error: Task not found: buidl.
	denoTaskNotFound = regexp.MustCompile(`Task not found: (\S+)`)
	// denoAnsiColor matches ANSI SGR sequences in the colored deno output.
	denoAnsiColor = regexp.MustCompile(`\x1b\[[0-9;]*m`)
)

// denoAvailableTasksHeading is printed before the list of available tasks.
const denoAvailableTasksHeading = "Available tasks"

// denoOutput strips ANSI color sequences from the captured deno output.
func denoOutput(command *types.Command) string {
	return denoAnsiColor.ReplaceAllString(command.Output, "")
}

// denoTaskNotFoundMatch reports whether deno failed because the task does not
// exist and listed the available tasks.
func denoTaskNotFoundMatch(command *types.Command) bool {
	output := denoOutput(command)
	if denoTaskNotFound.FindStringSubmatch(output) == nil {
		return false
	}
	return len(denoAvailableTasks(output)) > 0
}

// denoTaskNotFoundNewCommand replaces the unknown task with the tasks that are
// close to the misspelled one.
func denoTaskNotFoundNewCommand(command *types.Command) ([]string, error) {
	match := denoTaskNotFound.FindStringSubmatch(denoOutput(command))
	if match == nil {
		return nil, nil
	}
	return utils.ReplaceCommand(command, match[1], denoAvailableTasks(denoOutput(command))), nil
}

// denoAvailableTasks extracts the task names from the "Available tasks:" list
// of the deno output. Every task line starts with "- " and may be followed by
// a workspace marker such as "(workspace)".
func denoAvailableTasks(output string) []string {
	tasks := []string{}
	inTasks := false
	for _, line := range strings.Split(output, "\n") {
		trimmed := strings.TrimSpace(line)
		if !inTasks {
			inTasks = strings.HasPrefix(trimmed, denoAvailableTasksHeading)
			continue
		}
		if !strings.HasPrefix(trimmed, "- ") {
			continue
		}
		fields := strings.Fields(trimmed)
		if len(fields) >= 2 {
			tasks = append(tasks, fields[1])
		}
	}
	return tasks
}

func init() {
	register(types.NewRule(
		"deno_task_not_found",
		utils.ForApp([]string{"deno"}, 0, denoTaskNotFoundMatch),
		denoTaskNotFoundNewCommand,
	))
}
