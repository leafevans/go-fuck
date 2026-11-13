package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/v2/internal/specific"
	"github.com/leafevans/go-fuck/v2/internal/types"
	"github.com/leafevans/go-fuck/v2/internal/utils"
)

// npmPeerDependencyConflictError matches the error that npm prints when it
// cannot resolve a peer dependency tree.
const npmPeerDependencyConflictError = "ERESOLVE unable to resolve dependency tree"

// npmPeerDependencyConflictFix is the flag that npm itself recommends for the
// error. --force is not offered because it accepts a broken tree as well.
const npmPeerDependencyConflictFix = "--legacy-peer-deps"

// npmPeerDependencyConflictMatch reports whether npm failed to resolve a peer
// dependency tree. A command that already carries one of the escape hatches
// is not corrected again.
func npmPeerDependencyConflictMatch(command *types.Command) bool {
	if !strings.Contains(command.Output, npmPeerDependencyConflictError) {
		return false
	}
	for _, part := range command.ScriptParts() {
		if part == npmPeerDependencyConflictFix || part == "--force" {
			return false
		}
	}
	return true
}

// npmPeerDependencyConflictNewCommand retries the command with
// --legacy-peer-deps.
func npmPeerDependencyConflictNewCommand(command *types.Command) ([]string, error) {
	return []string{command.Script + " " + npmPeerDependencyConflictFix}, nil
}

func init() {
	// Python: enabled_by_default = npm_available
	rule := types.NewRule("npm_peer_dependency_conflict",
		utils.ForApp([]string{"npm"}, 0, npmPeerDependencyConflictMatch),
		npmPeerDependencyConflictNewCommand)
	rule.EnabledByDefault = specific.NpmAvailable()
	register(rule)
}
