package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/internal/shells"
	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// uvMissingProjectError is the error uv prints when no project file exists.
const uvMissingProjectError = "No `pyproject.toml` found in current directory or any parent directory"

// uvMissingProjectMatch reports whether uv failed because no project exists.
//
// uv pip does not need a project, so its subcommand is ignored even when the
// output contains the message.
func uvMissingProjectMatch(command *types.Command) bool {
	parts := command.ScriptParts()
	if len(parts) > 1 && strings.EqualFold(parts[1], "pip") {
		return false
	}
	return strings.Contains(command.Output, uvMissingProjectError)
}

// uvMissingProjectNewCommand initializes a project and retries.
func uvMissingProjectNewCommand(command *types.Command) ([]string, error) {
	return []string{shells.Current().And("uv init", command.Script)}, nil
}

func init() {
	register(types.NewRule("uv_missing_project",
		utils.ForApp([]string{"uv"}, 0, uvMissingProjectMatch),
		uvMissingProjectNewCommand))
}
