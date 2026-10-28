package rules

import (
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/leafevans/go-fuck/internal/shells"
	"github.com/leafevans/go-fuck/internal/system"
	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// pathFromHistoryPatterns matches the missing destination reported in the
// command output. The order is significant: only the first pattern that names
// an argument of the command is used.
var pathFromHistoryPatterns = []*regexp.Regexp{
	regexp.MustCompile(`no such file or directory: (.*)$`),
	regexp.MustCompile(`cannot access '(.*)': No such file or directory`),
	regexp.MustCompile(`: (.*): No such file or directory`),
	regexp.MustCompile(`can't cd to (.*)$`),
}

// pathFromHistoryGetValidHistory returns the history entries that may be
// suggested. It is a variable so that tests can inject a fixed history.
var pathFromHistoryGetValidHistory = utils.GetValidHistoryWithoutCurrent

// pathFromHistoryExists reports whether a path from the history exists. It is a
// variable so that tests do not depend on the local filesystem.
var pathFromHistoryExists = func(path string) bool {
	_, err := os.Stat(system.ExpandHome(path))
	return err == nil
}

// pathFromHistoryDestination returns the argument that caused the error, or an
// empty string when the output does not name one of the command arguments.
func pathFromHistoryDestination(command *types.Command) string {
	// Python's $ anchor matches at the end of the output or just before a
	// trailing newline, so the trailing newline is removed before searching.
	output := strings.TrimSuffix(command.Output, "\n")
	parts := command.ScriptParts()

	for _, pattern := range pathFromHistoryPatterns {
		match := pattern.FindStringSubmatch(output)
		if match == nil {
			continue
		}
		if utils.Contains(parts, match[1]) {
			return match[1]
		}
	}
	return ""
}

// pathFromHistoryAllAbsolutePaths returns the absolute paths used in the
// history, most used first.
func pathFromHistoryAllAbsolutePaths(command *types.Command) []string {
	counts := map[string]int{}
	order := []string{}

	for _, line := range pathFromHistoryGetValidHistory(command) {
		for _, param := range shells.Current().SplitCommand(line)[1:] {
			if !strings.HasPrefix(param, "/") && !strings.HasPrefix(param, "~") {
				continue
			}
			param = strings.TrimSuffix(param, "/")
			if _, seen := counts[param]; !seen {
				order = append(order, param)
			}
			counts[param]++
		}
	}

	sort.SliceStable(order, func(left, right int) bool {
		return counts[order[left]] > counts[order[right]]
	})
	return order
}

// pathFromHistoryNewCommand replaces the missing path with a matching absolute
// path used earlier in the history.
func pathFromHistoryNewCommand(command *types.Command) ([]string, error) {
	destination := pathFromHistoryDestination(command)
	if destination == "" {
		return nil, nil
	}

	results := []string{}
	for _, path := range pathFromHistoryAllAbsolutePaths(command) {
		if !strings.HasSuffix(path, destination) || !pathFromHistoryExists(path) {
			continue
		}
		results = append(results, utils.ReplaceArgument(command.Script, destination, path))
	}
	return results, nil
}

// pathFromHistoryMatch reports whether the output names a missing argument.
func pathFromHistoryMatch(command *types.Command) bool {
	return pathFromHistoryDestination(command) != ""
}

func init() {
	rule := types.NewRule("path_from_history", pathFromHistoryMatch, pathFromHistoryNewCommand)
	rule.Priority = 800
	register(rule)
}
