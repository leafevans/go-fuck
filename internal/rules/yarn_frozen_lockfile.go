package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/v2/internal/types"
	"github.com/leafevans/go-fuck/v2/internal/utils"
)

// yarnFrozenLockfileError matches the error that yarn prints when the lock file
// is stale and --frozen-lockfile is used.
const yarnFrozenLockfileError = "Your lockfile needs to be updated, but yarn was run with `--frozen-lockfile`"

// yarnFrozenLockfileMatch reports whether yarn refused to update a stale lock
// file.
func yarnFrozenLockfileMatch(command *types.Command) bool {
	return strings.Contains(command.Output, yarnFrozenLockfileError) &&
		utils.Contains(command.ScriptParts(), "yarn")
}

// yarnFrozenLockfileNewCommand overrides the frozen lock file and updates it.
func yarnFrozenLockfileNewCommand(command *types.Command) ([]string, error) {
	if strings.Contains(command.Script, "--frozen-lockfile") {
		return []string{strings.Replace(command.Script,
			"--frozen-lockfile", "--no-frozen-lockfile", 1)}, nil
	}
	return []string{command.Script + " --no-frozen-lockfile"}, nil
}

func init() {
	register(types.NewRule("yarn_frozen_lockfile",
		yarnFrozenLockfileMatch,
		yarnFrozenLockfileNewCommand))
}
