package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/v2/internal/types"
	"github.com/leafevans/go-fuck/v2/internal/utils"
)

// uvLockedLockfileError matches the error that uv prints when the lock file is
// stale and --locked is used.
const uvLockedLockfileError = "needs to be updated, but `--locked` was provided"

// uvLockedLockfileMatch reports whether uv refused to update a stale lock file.
func uvLockedLockfileMatch(command *types.Command) bool {
	return strings.Contains(command.Output, uvLockedLockfileError) &&
		utils.Contains(command.ScriptParts(), "--locked")
}

// uvLockedLockfileNewCommand drops --locked so that uv can update the lock
// file again.
func uvLockedLockfileNewCommand(command *types.Command) ([]string, error) {
	fixed := strings.TrimSpace(strings.Replace(command.Script, " --locked", "", 1))
	if fixed == "" || fixed == command.Script {
		return nil, nil
	}
	return []string{fixed}, nil
}

func init() {
	register(types.NewRule("uv_locked_lockfile",
		utils.ForApp([]string{"uv"}, 0, uvLockedLockfileMatch),
		uvLockedLockfileNewCommand))
}
