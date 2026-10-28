package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/internal/specific"
	"github.com/leafevans/go-fuck/internal/types"
)

// gitMainMasterMatch reports whether git failed because of the main or master
// branch.
func gitMainMasterMatch(command *types.Command) bool {
	return strings.Contains(command.Output, "'master'") ||
		strings.Contains(command.Output, "'main'")
}

// gitMainMasterNewCommand switches between the main and master branch names.
func gitMainMasterNewCommand(command *types.Command) ([]string, error) {
	if strings.Contains(command.Output, "'master'") {
		return []string{strings.ReplaceAll(command.Script, "master", "main")}, nil
	}
	return []string{strings.ReplaceAll(command.Script, "main", "master")}, nil
}

func init() {
	rule := types.NewRule("git_main_master",
		specific.GitSupportMatch(gitMainMasterMatch),
		specific.GitSupportNew(gitMainMasterNewCommand))
	rule.Priority = 1200
	register(rule)
}
