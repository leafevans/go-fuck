package rules

import (
	"os"
	"strings"

	"github.com/leafevans/go-fuck/internal/specific"
	"github.com/leafevans/go-fuck/internal/types"
)

// hasExistsScriptPathExists reports whether a path exists. It is a variable so
// that tests do not depend on the local filesystem.
var hasExistsScriptPathExists = func(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// hasExistsScriptMatch reports whether the executable exists as a local file.
func hasExistsScriptMatch(command *types.Command) bool {
	parts := command.ScriptParts()
	if len(parts) == 0 {
		return false
	}
	return hasExistsScriptPathExists(parts[0]) &&
		strings.Contains(command.Output, "command not found")
}

// hasExistsScriptNewCommand runs the local file explicitly.
func hasExistsScriptNewCommand(command *types.Command) ([]string, error) {
	return []string{"./" + command.Script}, nil
}

func init() {
	register(types.NewRule("has_exists_script",
		specific.SudoSupportMatch(hasExistsScriptMatch),
		specific.SudoSupportNew(hasExistsScriptNewCommand)))
}
