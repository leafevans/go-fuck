package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/v2/internal/shells"
	"github.com/leafevans/go-fuck/v2/internal/specific"
	"github.com/leafevans/go-fuck/v2/internal/types"
	"github.com/leafevans/go-fuck/v2/internal/utils"
)

// gitRemoteAlreadyExistsError matches the error that git prints when
// "git remote add" is used for a remote that already exists.
var gitRemoteAlreadyExistsError = newRegexp(`(?m)^(?:error|fatal): remote (\S+) already exists\.$`)

// gitRemoteAlreadyExistsMatch reports whether "git remote add" failed because
// the remote already exists. Flags between "add" and the remote name are not
// supported, because "git remote set-url" does not accept the same flags.
func gitRemoteAlreadyExistsMatch(command *types.Command) bool {
	_, ok := gitRemoteAlreadyExistsCorrection(command)
	return ok
}

// gitRemoteAlreadyExistsNewCommand changes the remote instead of trying to add
// it again.
func gitRemoteAlreadyExistsNewCommand(command *types.Command) ([]string, error) {
	corrected, ok := gitRemoteAlreadyExistsCorrection(command)
	if !ok {
		return nil, nil
	}
	return []string{corrected}, nil
}

func gitRemoteAlreadyExistsCorrection(command *types.Command) (string, bool) {
	if !gitRemoteAlreadyExistsError.MatchString(command.Output) {
		return "", false
	}
	operation, _, ok := gitOperationWordSpan(command.Script, "remote")
	if !ok {
		return "", false
	}
	for span, word := range specific.GitCommandWords(command.Script, shells.Current()) {
		if span[0] <= operation[1] {
			continue
		}
		if strings.HasPrefix(word, "-") {
			return "", false
		}
		break
	}
	return gitReplaceOperation(command.Script, "remote", "add", "set-url")
}

func init() {
	register(types.NewRule("git_remote_already_exists",
		utils.ForApp([]string{"git"}, 0, gitRemoteAlreadyExistsMatch),
		gitRemoteAlreadyExistsNewCommand))
}
