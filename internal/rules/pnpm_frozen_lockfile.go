package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/v2/internal/types"
	"github.com/leafevans/go-fuck/v2/internal/utils"
)

// pnpmFrozenLockfileError matches the error that pnpm prints when the lock
// file is stale and --frozen-lockfile is used.
const pnpmFrozenLockfileError = "ERR_PNPM_OUTDATED_LOCKFILE"

// pnpmFrozenLockfileMatch reports whether pnpm refused to update a stale lock
// file.
func pnpmFrozenLockfileMatch(command *types.Command) bool {
	return strings.Contains(command.Output, pnpmFrozenLockfileError) &&
		utils.Contains(command.ScriptParts(), "pnpm")
}

// pnpmFrozenLockfileNewCommand overrides the frozen lock file and updates it.
func pnpmFrozenLockfileNewCommand(command *types.Command) ([]string, error) {
	if strings.Contains(command.Script, "--frozen-lockfile") {
		return []string{strings.Replace(command.Script,
			"--frozen-lockfile", "--no-frozen-lockfile", 1)}, nil
	}
	return []string{command.Script + " --no-frozen-lockfile"}, nil
}

func init() {
	register(types.NewRule("pnpm_frozen_lockfile",
		pnpmFrozenLockfileMatch,
		pnpmFrozenLockfileNewCommand))
}
