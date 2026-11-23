package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/v3/internal/shells"
	"github.com/leafevans/go-fuck/v3/internal/specific"
	"github.com/leafevans/go-fuck/v3/internal/types"
	"github.com/leafevans/go-fuck/v3/internal/utils"
)

// gitRemoteSeturlAddMatch reports whether git remote set-url was used on a
// remote that does not exist yet.
func gitRemoteSeturlAddMatch(command *types.Command) bool {
	if !strings.Contains(command.Output, "No such remote") {
		return false
	}
	if _, ok := gitRemoteSeturlSpan(command.Script); !ok {
		return false
	}
	// When another configured remote is close to the unknown name, the name was
	// mistyped: git_unknown_remote corrects it instead of adding a new remote.
	if missing, ok := gitUnknownRemoteName(command.Output); ok && missing != "" {
		closest := utils.GetClosest(missing, gitRemoteNamesCommand(command.Script), 0.6, false)
		if closest != "" && closest != missing {
			return false
		}
	}
	return true
}

// gitRemoteSeturlAddNewCommand adds the remote instead of updating it.
func gitRemoteSeturlAddNewCommand(command *types.Command) ([]string, error) {
	span, ok := gitRemoteSeturlSpan(command.Script)
	if !ok {
		return nil, nil
	}
	return []string{command.Script[:span[0]] + "add" + command.Script[span[1]:]}, nil
}

// Only the plain set-url <remote> <url> form has an equivalent remote add.
func gitRemoteSeturlSpan(script string) ([2]int, bool) {
	var operation [2]int
	count := 0
	for span, word := range specific.GitCommandPathWords(script, shells.Current()) {
		switch count {
		case 0:
			if word != "remote" {
				return operation, false
			}
		case 1:
			if word != "set-url" {
				return operation, false
			}
			operation = span
		case 2, 3:
			if word == "" || strings.HasPrefix(word, "-") {
				return operation, false
			}
		default:
			return operation, false
		}
		count++
	}
	return operation, count == 4
}

func init() {
	register(types.NewRule("git_remote_seturl_add",
		specific.GitSupportMatch(gitRemoteSeturlAddMatch),
		specific.GitSupportNew(gitRemoteSeturlAddNewCommand)))
}
