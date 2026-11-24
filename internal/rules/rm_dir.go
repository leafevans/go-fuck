package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/v3/internal/specific"
	"github.com/leafevans/go-fuck/v3/internal/types"
)

// rmDirPattern matches a rm invocation.
var rmDirPattern = newRegexp(`\brm (.*)`)

// rmDirMatch reports whether rm failed because the target is a directory.
//
// Deviation: the reference implementation matches any script that contains the
// word rm, so "grep -rn rm build" failing with "Is a directory" was rewritten
// into "grep -rn rm -rf build". The word has to be an actual invocation before
// the recursive removal is suggested.
func rmDirMatch(command *types.Command) bool {
	return commandInvocation(command.Script, "rm") &&
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
