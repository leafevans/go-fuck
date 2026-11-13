package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/v2/internal/types"
)

// cmdRmdirNotEmptyError is the error that cmd.exe prints for rmdir on a
// directory with children.
const cmdRmdirNotEmptyError = "The directory is not empty."

// cmdRmdirNonEmptyMatch reports whether rmdir failed on a non-empty directory.
func cmdRmdirNonEmptyMatch(command *types.Command) bool {
	parts := command.ScriptParts()
	if len(parts) == 0 {
		return false
	}
	if !strings.EqualFold(parts[0], "rmdir") && !strings.EqualFold(parts[0], "rd") {
		return false
	}
	return strings.Contains(command.Output, cmdRmdirNotEmptyError)
}

// cmdRmdirNonEmptyNewCommand removes the directory tree.
func cmdRmdirNonEmptyNewCommand(command *types.Command) ([]string, error) {
	return []string{cmdInsertArgument(command.Script, "/s /q")}, nil
}

func init() {
	register(types.NewRule("cmd_rmdir_nonempty",
		cmdRmdirNonEmptyMatch,
		cmdRmdirNonEmptyNewCommand))
}
