package rules

import (
	"regexp"
	"strings"

	"github.com/leafevans/go-fuck/internal/shells"
	"github.com/leafevans/go-fuck/internal/specific"
	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// cdMkdirScript matches the directory argument of a cd command.
var cdMkdirScript = regexp.MustCompile("^cd (.*)")

// cdMkdirMatch reports whether cd failed because the directory is missing.
func cdMkdirMatch(command *types.Command) bool {
	if !strings.HasPrefix(command.Script, "cd ") {
		return false
	}
	output := strings.ToLower(command.Output)
	return strings.Contains(output, "no such file or directory") ||
		strings.Contains(output, "cd: can't cd to") ||
		strings.Contains(output, "does not exist")
}

// cdMkdirNewCommand creates the directory and changes into it.
func cdMkdirNewCommand(command *types.Command) ([]string, error) {
	replacement := shells.Current().And("mkdir -p $1", "cd $1")
	return []string{cdMkdirScript.ReplaceAllString(command.Script, replacement)}, nil
}

func init() {
	register(types.NewRule("cd_mkdir",
		specific.SudoSupportMatch(utils.ForApp([]string{"cd"}, 0, cdMkdirMatch)),
		specific.SudoSupportNew(cdMkdirNewCommand)))
}
