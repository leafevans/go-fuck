package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/internal/specific"
	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// gitPushForceMatch reports whether a push was rejected because the remote
// branch is ahead.
func gitPushForceMatch(command *types.Command) bool {
	return strings.Contains(command.Script, "push") &&
		strings.Contains(command.Output, "! [rejected]") &&
		strings.Contains(command.Output, "failed to push some refs to") &&
		strings.Contains(command.Output,
			"Updates were rejected because the tip of your current branch is behind")
}

// gitPushForceNewCommand forces the push with a lease.
func gitPushForceNewCommand(command *types.Command) ([]string, error) {
	return []string{utils.ReplaceArgument(command.Script, "push", "push --force-with-lease")}, nil
}

func init() {
	rule := types.NewRule("git_push_force",
		specific.GitSupportMatch(gitPushForceMatch),
		specific.GitSupportNew(gitPushForceNewCommand))
	// Python: enabled_by_default = False
	rule.EnabledByDefault = false
	register(rule)
}
