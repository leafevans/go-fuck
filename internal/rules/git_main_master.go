package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/v2/internal/shells"
	"github.com/leafevans/go-fuck/v2/internal/specific"
	"github.com/leafevans/go-fuck/v2/internal/types"
)

// gitMainMasterMatch reports whether git failed because of the main or master
// branch.
func gitMainMasterMatch(command *types.Command) bool {
	_, ok := gitMainMasterCorrection(command)
	return ok
}

// gitMainMasterNewCommand switches between the main and master branch names.
func gitMainMasterNewCommand(command *types.Command) ([]string, error) {
	corrected, ok := gitMainMasterCorrection(command)
	if !ok {
		return nil, nil
	}
	return []string{corrected}, nil
}

func gitMainMasterCorrection(command *types.Command) (string, bool) {
	from, to := "main", "master"
	hasMaster := strings.Contains(command.Output, "'master'") || strings.TrimSpace(command.Output) == "fatal: invalid reference: master"
	hasMain := strings.Contains(command.Output, "'main'") || strings.TrimSpace(command.Output) == "fatal: invalid reference: main"
	if hasMaster == hasMain {
		return "", false
	}
	if hasMaster {
		from, to = "master", "main"
	}
	if !strings.Contains(command.Output, "bad revision '"+from+"'") &&
		!strings.Contains(command.Output, "pathspec '"+from+"' did not match") &&
		!strings.Contains(command.Output, "ambiguous argument '"+from+"': unknown revision or path") &&
		!strings.Contains(command.Output, "invalid reference: "+from) {
		return "", false
	}
	start, end, ok := specific.GitArgumentSpan(command.Script, from, shells.Current(), false)
	if !ok {
		return "", false
	}
	previous := ""
	for span, word := range specific.GitCommandWords(command.Script, shells.Current()) {
		if previous == "" {
			switch word {
			case "checkout", "switch", "log", "show", "diff", "merge", "rebase", "rev-parse":
			default:
				return "", false
			}
		}
		if span[0] == start {
			// An option immediately before the name may consume it as a value.
			if strings.HasPrefix(previous, "-") && !strings.Contains(previous, "=") &&
				strings.Trim(previous[1:], "0123456789") != "" {
				return "", false
			}
			break
		}
		previous = word
	}
	return command.Script[:start] + to + command.Script[end:], true
}

func init() {
	rule := types.NewRule("git_main_master",
		specific.GitSupportMatch(gitMainMasterMatch),
		specific.GitSupportNew(gitMainMasterNewCommand))
	rule.Priority = 1200
	register(rule)
}
