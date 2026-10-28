package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/internal/specific"
	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// gitRebaseNoChangesMatch reports whether git rebase --continue was run after
// its changes were already applied.
func gitRebaseNoChangesMatch(command *types.Command) bool {
	parts := command.ScriptParts()
	return utils.Contains(parts, "rebase") &&
		utils.Contains(parts, "--continue") &&
		strings.Contains(command.Output,
			"No changes - did you forget to use 'git add'?")
}

// gitRebaseNoChangesNewCommand skips the empty patch.
func gitRebaseNoChangesNewCommand(command *types.Command) ([]string, error) {
	return []string{"git rebase --skip"}, nil
}

func init() {
	register(types.NewRule("git_rebase_no_changes",
		specific.GitSupportMatch(gitRebaseNoChangesMatch),
		gitRebaseNoChangesNewCommand))
}
