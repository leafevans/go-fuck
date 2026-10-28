package rules

import (
	"regexp"
	"strings"

	"github.com/leafevans/go-fuck/internal/specific"
	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// cpOmittingDirectoryPrefix matches the leading cp command.
var cpOmittingDirectoryPrefix = regexp.MustCompile("^cp")

// cpOmittingDirectoryMatch reports whether cp rejected a directory argument.
func cpOmittingDirectoryMatch(command *types.Command) bool {
	output := strings.ToLower(command.Output)
	return strings.Contains(output, "omitting directory") ||
		strings.Contains(output, "is a directory")
}

// cpOmittingDirectoryNewCommand makes cp copy directories recursively.
func cpOmittingDirectoryNewCommand(command *types.Command) ([]string, error) {
	script := cpOmittingDirectoryPrefix.ReplaceAllString(command.Script, "cp -a")
	return []string{script}, nil
}

func init() {
	register(types.NewRule("cp_omitting_directory",
		specific.SudoSupportMatch(utils.ForApp([]string{"cp"}, 0, cpOmittingDirectoryMatch)),
		specific.SudoSupportNew(cpOmittingDirectoryNewCommand)))
}
