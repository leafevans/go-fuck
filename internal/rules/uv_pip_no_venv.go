package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/v2/internal/shells"
	"github.com/leafevans/go-fuck/v2/internal/types"
	"github.com/leafevans/go-fuck/v2/internal/utils"
)

// uvNoVirtualEnvError is the error uv pip prints when no virtual environment
// exists.
const uvNoVirtualEnvError = "No virtual environment found"

// uvPipNoVenvMatch reports whether uv pip failed because no virtual
// environment exists.
func uvPipNoVenvMatch(command *types.Command) bool {
	parts := command.ScriptParts()
	if len(parts) < 2 || !strings.EqualFold(parts[1], "pip") {
		return false
	}
	return strings.Contains(command.Output, uvNoVirtualEnvError)
}

// uvPipNoVenvNewCommand creates the virtual environment and retries.
func uvPipNoVenvNewCommand(command *types.Command) ([]string, error) {
	return []string{shells.Current().And("uv venv", command.Script)}, nil
}

func init() {
	register(types.NewRule("uv_pip_no_venv",
		utils.ForApp([]string{"uv"}, 1, uvPipNoVenvMatch),
		uvPipNoVenvNewCommand))
}
