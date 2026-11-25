package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/v4/internal/shells"
	"github.com/leafevans/go-fuck/v4/internal/types"
)

// noSuchFilePatterns match the mv and cp errors that are caused by a missing
// destination directory.
var noSuchFilePatterns = []*lazyRegexp{
	newRegexp("mv: cannot move '[^']*' to '([^']*)': No such file or directory"),
	newRegexp("mv: cannot move '[^']*' to '([^']*)': Not a directory"),
	newRegexp("cp: cannot create regular file '([^']*)': No such file or directory"),
	newRegexp("cp: cannot create regular file '([^']*)': Not a directory"),
}

// noSuchFileMatch reports whether mv or cp failed because the destination
// directory is missing.
func noSuchFileMatch(command *types.Command) bool {
	// Every pattern names one of these two messages; checking them as plain
	// literals keeps a long capture away from four full scans.
	if !strings.Contains(command.Output, "No such file or directory") &&
		!strings.Contains(command.Output, "Not a directory") {
		return false
	}
	for _, pattern := range noSuchFilePatterns {
		if pattern.MatchString(command.Output) {
			return true
		}
	}
	return false
}

// noSuchFileNewCommand creates the missing destination directory first.
func noSuchFileNewCommand(command *types.Command) ([]string, error) {
	for _, pattern := range noSuchFilePatterns {
		matches := pattern.FindAllStringSubmatch(command.Output, -1)
		if len(matches) == 0 {
			continue
		}
		file := matches[0][1]

		// Python: dir = file[0:file.rfind('/')]. A file name without a slash
		// makes rfind return -1, which drops the last character.
		directory := ""
		if before, _, found := strings.CutLast(file, "/"); found {
			directory = before
		} else if file != "" {
			directory = file[:len(file)-1]
		}
		return []string{shells.Current().And("mkdir -p "+directory, command.Script)}, nil
	}
	return nil, nil
}

func init() {
	register(types.NewRule("no_such_file", noSuchFileMatch, noSuchFileNewCommand))
}
