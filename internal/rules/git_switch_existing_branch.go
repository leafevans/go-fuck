package rules

import (
	"regexp"
	"strings"

	"github.com/leafevans/go-fuck/internal/specific"
	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// gitSwitchExistingBranchName matches the branch name that already exists.
var gitSwitchExistingBranchName = regexp.MustCompile(`fatal: a branch named '([^']*)' already exists`)

// gitSwitchExistingBranchMatch reports whether git switch failed because the
// branch it should create already exists.
func gitSwitchExistingBranchMatch(command *types.Command) bool {
	parts := command.ScriptParts()
	if !utils.Contains(parts, "switch") ||
		(!utils.Contains(parts, "-c") && !utils.Contains(parts, "--create")) {
		return false
	}
	return gitSwitchExistingBranchName.MatchString(command.Output)
}

// gitSwitchExistingBranchNewCommand switches to the existing branch instead of
// creating it.
func gitSwitchExistingBranchNewCommand(command *types.Command) ([]string, error) {
	if !gitSwitchExistingBranchName.MatchString(command.Output) {
		return nil, nil
	}
	fixed := strings.Replace(command.Script, " --create ", " ", 1)
	fixed = strings.TrimSpace(strings.Replace(fixed, " -c ", " ", 1))
	if fixed == "" || fixed == command.Script {
		return nil, nil
	}
	return []string{fixed}, nil
}

func init() {
	register(types.NewRule("git_switch_existing_branch",
		specific.GitSupportMatch(gitSwitchExistingBranchMatch),
		specific.GitSupportNew(gitSwitchExistingBranchNewCommand)))
}
