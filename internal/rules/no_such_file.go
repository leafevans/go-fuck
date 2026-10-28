package rules

import (
	"regexp"
	"strings"

	"github.com/leafevans/go-fuck/internal/shells"
	"github.com/leafevans/go-fuck/internal/types"
)

// noSuchFilePatterns match the mv and cp errors that are caused by a missing
// destination directory.
var noSuchFilePatterns = []*regexp.Regexp{
	regexp.MustCompile("mv: cannot move '[^']*' to '([^']*)': No such file or directory"),
	regexp.MustCompile("mv: cannot move '[^']*' to '([^']*)': Not a directory"),
	regexp.MustCompile("cp: cannot create regular file '([^']*)': No such file or directory"),
	regexp.MustCompile("cp: cannot create regular file '([^']*)': Not a directory"),
}

// noSuchFileMatch reports whether mv or cp failed because the destination
// directory is missing.
func noSuchFileMatch(command *types.Command) bool {
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
		if index := strings.LastIndex(file, "/"); index >= 0 {
			directory = file[:index]
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
