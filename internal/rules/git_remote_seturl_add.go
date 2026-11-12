package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/internal/specific"
	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// gitRemoteSeturlAddMatch reports whether git remote set-url was used on a
// remote that does not exist yet.
func gitRemoteSeturlAddMatch(command *types.Command) bool {
	if !strings.Contains(command.Script, "set-url") ||
		!strings.Contains(command.Output, "No such remote") {
		return false
	}
	// When another configured remote is close to the unknown name, the name was
	// mistyped: git_unknown_remote corrects it instead of adding a new remote.
	if missing, ok := gitUnknownRemoteName(command.Output); ok && missing != "" {
		closest := utils.GetClosest(missing, gitRemoteNamesCommand(), 0.6, false)
		if closest != "" && closest != missing {
			return false
		}
	}
	return true
}

// gitRemoteSeturlAddNewCommand adds the remote instead of updating it.
func gitRemoteSeturlAddNewCommand(command *types.Command) ([]string, error) {
	return []string{utils.ReplaceArgument(command.Script, "set-url", "add")}, nil
}

func init() {
	register(types.NewRule("git_remote_seturl_add",
		specific.GitSupportMatch(gitRemoteSeturlAddMatch),
		gitRemoteSeturlAddNewCommand))
}
