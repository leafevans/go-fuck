package rules

import (
	"os"
	"regexp"
	"strings"

	"github.com/leafevans/go-fuck/internal/shells"
	"github.com/leafevans/go-fuck/internal/specific"
	"github.com/leafevans/go-fuck/internal/types"
)

// gitAddPathspec matches the path that git did not find.
var gitAddPathspec = regexp.MustCompile(
	`error: pathspec '([^']*)' did not match any file\(s\) known to git\.`)

// gitAddFileExists reports whether a path exists. It is a variable so that
// tests do not depend on the file system.
var gitAddFileExists = func(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// gitAddMissingFile returns the existing path that git failed to find, or an
// empty string when there is none.
func gitAddMissingFile(command *types.Command) string {
	match := gitAddPathspec.FindStringSubmatch(command.Output)
	if match == nil || !gitAddFileExists(match[1]) {
		return ""
	}
	return match[1]
}

// gitAddMatch reports whether git failed because of an unknown pathspec that
// exists on disk.
func gitAddMatch(command *types.Command) bool {
	return strings.Contains(command.Output, "did not match any file(s) known to git.") &&
		gitAddMissingFile(command) != ""
}

// gitAddNewCommand adds the missing file first.
func gitAddNewCommand(command *types.Command) ([]string, error) {
	missingFile := gitAddMissingFile(command)
	if missingFile == "" {
		return nil, nil
	}
	return []string{shells.Current().And("git add -- "+missingFile, command.Script)}, nil
}

func init() {
	register(types.NewRule("git_add",
		specific.GitSupportMatch(gitAddMatch),
		specific.GitSupportNew(gitAddNewCommand)))
}
