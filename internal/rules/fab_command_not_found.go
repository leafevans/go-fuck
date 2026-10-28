package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// fabCommandNotFoundGetBetween returns the first word of every non-empty line
// between start and end. When end is empty the lines after start are used.
func fabCommandNotFoundGetBetween(content, start, end string) []string {
	results := []string{}
	shouldYield := false
	for _, line := range strings.Split(content, newline) {
		if strings.Contains(line, start) {
			shouldYield = true
			continue
		}
		if end != "" && strings.Contains(line, end) {
			return results
		}
		if shouldYield && line != "" {
			results = append(results, strings.Split(strings.TrimSpace(line), " ")[0])
		}
	}
	return results
}

// fabCommandNotFoundMatch reports whether fab did not find a task.
func fabCommandNotFoundMatch(command *types.Command) bool {
	return strings.Contains(command.Output, "Warning: Command(s) not found:")
}

// fabCommandNotFoundNewCommand replaces the unknown tasks with the closest
// known ones.
func fabCommandNotFoundNewCommand(command *types.Command) ([]string, error) {
	notFound := fabCommandNotFoundGetBetween(
		command.Output, "Warning: Command(s) not found:", "Available commands:")
	possible := fabCommandNotFoundGetBetween(
		command.Output, "Available commands:", "")

	script := command.Script
	for _, missing := range notFound {
		fix := utils.GetClosestDefault(missing, possible)
		script = strings.ReplaceAll(script, " "+missing, " "+fix)
	}
	return []string{script}, nil
}

func init() {
	register(types.NewRule("fab_command_not_found",
		utils.ForApp([]string{"fab"}, 0, fabCommandNotFoundMatch),
		fabCommandNotFoundNewCommand))
}
