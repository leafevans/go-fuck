package rules

import (
	"regexp"
	"strings"

	"github.com/leafevans/go-fuck/internal/specific"
	"github.com/leafevans/go-fuck/internal/types"
)

// mkdirPPattern matches a mkdir invocation.
var mkdirPPattern = regexp.MustCompile(`\bmkdir (.*)`)

// mkdirPMatch reports whether mkdir failed because the parent directory is
// missing.
func mkdirPMatch(command *types.Command) bool {
	return strings.Contains(command.Script, "mkdir") &&
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
