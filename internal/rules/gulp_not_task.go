package rules

import (
	"os/exec"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// gulpNotTaskPattern matches the unknown gulp task reported by gulp.
var gulpNotTaskPattern = regexp.MustCompile(`Task '(\w+)' is not in your gulpfile`)

// gulpNotTaskParseTasks extracts the tasks from "gulp --tasks-simple" output.
//
// The original implementation strips exactly one trailing newline character of
// every line, so no other whitespace is removed.
func gulpNotTaskParseTasks(tasksOutput string) []string {
	lines := strings.Split(tasksOutput, "\n")
	tasks := make([]string, 0, len(lines))
	for index, line := range lines {
		if index == len(lines)-1 {
			if line == "" {
				continue
			}
			// The output does not end with a newline, so the original
			// implementation drops the last character of the last line.
			_, size := utf8.DecodeLastRuneInString(line)
			line = line[:len(line)-size]
		}
		tasks = append(tasks, line)
	}
	return tasks
}

// gulpTasks returns the gulp tasks of the current project.
//
// The result is memoized because gulp is started once per go-fuck invocation.
var gulpTasks = utils.Memoize(func() []string {
	output, err := exec.Command("gulp", "--tasks-simple").Output()
	if err != nil && len(output) == 0 {
		return nil
	}
	return gulpNotTaskParseTasks(string(output))
})

// gulpNotTaskMatch reports whether gulp reported an unknown task.
func gulpNotTaskMatch(command *types.Command) bool {
	return strings.Contains(command.Output, "is not in your gulpfile")
}

// gulpNotTaskNewCommand replaces the unknown task with the closest task.
func gulpNotTaskNewCommand(command *types.Command) ([]string, error) {
	wrongTask := gulpNotTaskPattern.FindStringSubmatch(command.Output)
	if wrongTask == nil {
		return nil, nil
	}
	return utils.ReplaceCommand(command, wrongTask[1], gulpTasks()), nil
}

func init() {
	register(types.NewRule("gulp_not_task",
		utils.ForApp([]string{"gulp"}, 0, gulpNotTaskMatch),
		gulpNotTaskNewCommand))
}
