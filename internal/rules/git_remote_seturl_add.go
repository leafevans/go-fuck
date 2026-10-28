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
	return strings.Contains(command.Script, "set-url") &&
		strings.Contains(command.Output, "fatal: No such remote")
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
