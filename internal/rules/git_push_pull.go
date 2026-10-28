package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/internal/shells"
	"github.com/leafevans/go-fuck/internal/specific"
	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// gitPushPullMatch reports whether a push was rejected because the remote
// contains work that is not integrated locally.
func gitPushPullMatch(command *types.Command) bool {
	return strings.Contains(command.Script, "push") &&
		strings.Contains(command.Output, "! [rejected]") &&
		strings.Contains(command.Output, "failed to push some refs to") &&
		(strings.Contains(command.Output,
			"Updates were rejected because the tip of your current branch is behind") ||
			strings.Contains(command.Output,
				"Updates were rejected because the remote contains work that you do"))
}

// gitPushPullNewCommand pulls the remote changes and retries the push.
func gitPushPullNewCommand(command *types.Command) ([]string, error) {
	pull := utils.ReplaceArgument(command.Script, "push", "pull")
	return []string{shells.Current().And(pull, command.Script)}, nil
}

func init() {
	register(types.NewRule("git_push_pull",
		specific.GitSupportMatch(gitPushPullMatch),
		specific.GitSupportNew(gitPushPullNewCommand)))
}
