package rules

import (
	"path/filepath"
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
	return mkdirPInvocation(command.Script) &&
		strings.Contains(command.Output, "No such file or directory")
}

// mkdirPInvocation reports whether script runs mkdir, either as the command
// itself or as the "-mkdir" subcommand of hdfs.
func mkdirPInvocation(script string) bool {
	fields := strings.Fields(script)
	if len(fields) == 0 {
		return false
	}
	if filepath.Base(fields[0]) == "mkdir" {
		return true
	}
	for index := 1; index < len(fields); index++ {
		if fields[index] != "-mkdir" {
			continue
		}
		switch filepath.Base(fields[index-1]) {
		case "hdfs", "dfs":
			return true
		}
	}
	return false
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
