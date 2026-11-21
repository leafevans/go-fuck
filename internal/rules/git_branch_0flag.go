package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/v2/internal/shells"
	"github.com/leafevans/go-fuck/v2/internal/specific"
	"github.com/leafevans/go-fuck/v2/internal/types"
)

// gitBranch0flagMatch reports whether a branch flag was mistyped with a zero.
func gitBranch0flagMatch(command *types.Command) bool {
	_, ok := gitBranch0flagCorrection(command)
	return ok
}

// gitBranch0flagNewCommand replaces the zero with a dash.
func gitBranch0flagNewCommand(command *types.Command) ([]string, error) {
	corrected, ok := gitBranch0flagCorrection(command)
	if !ok {
		return nil, nil
	}
	return []string{corrected}, nil
}

func gitBranch0flagCorrection(command *types.Command) (string, bool) {
	first := true
	for span, word := range specific.GitCommandWords(command.Script, shells.Current()) {
		if first {
			if word != "branch" {
				return "", false
			}
			first = false
			continue
		}
		if len(word) != 2 || word[0] != '0' || !strings.ContainsRune("adDflruv", rune(word[1])) {
			return "", false
		}
		if !strings.Contains(command.Output, "A branch named '"+word+"' already exists") &&
			!strings.Contains(command.Output, "Not a valid object name:") {
			return "", false
		}
		return command.Script[:span[0]] + "-" + word[1:] + command.Script[span[1]:], true
	}
	return "", false
}

func init() {
	register(types.NewRule("git_branch_0flag",
		specific.GitSupportMatch(gitBranch0flagMatch),
		specific.GitSupportNew(gitBranch0flagNewCommand)))
}
