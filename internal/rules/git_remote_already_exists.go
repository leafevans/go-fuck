package rules

import (
	"regexp"
	"strings"

	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// gitRemoteAlreadyExistsError matches the error that git prints when
// "git remote add" is used for a remote that already exists.
var gitRemoteAlreadyExistsError = regexp.MustCompile(`(?m)^(?:error|fatal): remote (\S+) already exists\.$`)

// gitRemoteAlreadyExistsMatch reports whether "git remote add" failed because
// the remote already exists. Flags between "add" and the remote name are not
// supported, because "git remote set-url" does not accept the same flags.
func gitRemoteAlreadyExistsMatch(command *types.Command) bool {
	return strings.Contains(command.Script, "remote add ") &&
		!strings.Contains(command.Script, "remote add -") &&
		gitRemoteAlreadyExistsError.MatchString(command.Output)
}

// gitRemoteAlreadyExistsNewCommand changes the remote instead of trying to add
// it again.
func gitRemoteAlreadyExistsNewCommand(command *types.Command) ([]string, error) {
	if !gitRemoteAlreadyExistsMatch(command) {
		return nil, nil
	}
	corrected := utils.ReplaceArgument(command.Script, "add", "set-url")
	if corrected == command.Script {
		return nil, nil
	}
	return []string{corrected}, nil
}

func init() {
	register(types.NewRule("git_remote_already_exists",
		utils.ForApp([]string{"git"}, 0, gitRemoteAlreadyExistsMatch),
		gitRemoteAlreadyExistsNewCommand))
}
