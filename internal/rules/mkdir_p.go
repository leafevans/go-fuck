package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/v3/internal/specific"
	"github.com/leafevans/go-fuck/v3/internal/types"
)

// mkdirPPattern matches a mkdir invocation.
var mkdirPPattern = newRegexp(`\bmkdir (.*)`)

// mkdirPMatch reports whether mkdir failed because the parent directory is
// missing.
//
// Deviation: the reference implementation matches any script that contains the
// word mkdir, so a command that only mentions mkdir in its arguments, such as
// "grep -rn mkdir src", was rewritten into "grep -rn mkdir -p src". The word has
// to be an actual invocation before the parent directory is created.
func mkdirPMatch(command *types.Command) bool {
	return commandInvocation(command.Script, "mkdir") &&
		strings.Contains(command.Output, "No such file or directory")
}

// mkdirPNewCommand creates the missing parent directories.
func mkdirPNewCommand(command *types.Command) ([]string, error) {
	return []string{mkdirPPattern.ReplaceAllString(command.Script, "mkdir -p $1")}, nil
}

func init() {
	register(types.NewRule("mkdir_p",
		specific.SudoSupportMatch(mkdirPMatch),
		specific.SudoSupportNew(mkdirPNewCommand)))
}
