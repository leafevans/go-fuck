package rules

import (
	"regexp"
	"strings"

	"github.com/leafevans/go-fuck/internal/specific"
	"github.com/leafevans/go-fuck/internal/types"
)

// rmDirPattern matches a rm invocation.
var rmDirPattern = regexp.MustCompile(`\brm (.*)`)

// rmDirMatch reports whether rm failed because the target is a directory.
func rmDirMatch(command *types.Command) bool {
	return strings.Contains(command.Script, "rm") &&
		strings.Contains(strings.ToLower(command.Output), "is a directory")
}

// rmDirNewCommand makes rm remove the directory recursively.
func rmDirNewCommand(command *types.Command) ([]string, error) {
	arguments := "-rf"
	if strings.Contains(command.Script, "hdfs") {
		arguments = "-r"
	}
	script := rmDirPattern.ReplaceAllString(command.Script, "rm "+arguments+" $1")
	return []string{script}, nil
}

func init() {
	register(types.NewRule("rm_dir",
		specific.SudoSupportMatch(rmDirMatch),
		specific.SudoSupportNew(rmDirNewCommand)))
}
