package rules

import (
	"os/exec"
	"regexp"
	"strings"

	"github.com/leafevans/go-fuck/v2/internal/types"
	"github.com/leafevans/go-fuck/v2/internal/utils"
)

// gitBranchMissingPatterns match the errors that git prints when a branch
// passed to "git branch" does not exist.
var gitBranchMissingPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?m)^error: branch '([^']+)' not found$`),
	regexp.MustCompile(`(?m)^fatal: no branch named '([^']+)'$`),
}

// gitLocalBranchesCommand lists the local branch names. It is a variable so
// that tests do not depend on the repository that runs them.
var gitLocalBranchesCommand = func() []string {
	output, err := exec.Command("git", "branch", "--list", "--format=%(refname:short)").Output()
	if err != nil {
		return nil
	}
	return strings.Fields(string(output))
}

// gitBranchMissingMatch reports whether "git branch" failed because the branch
// does not exist and another local branch is close to it.
func gitBranchMissingMatch(command *types.Command) bool {
	if !gitBranchCommandUsesName(command.Script) {
		return false
	}
	missing, ok := gitBranchMissingName(command.Output)
	if !ok || missing == "" {
		return false
	}
	closest := utils.GetClosest(missing, gitLocalBranchesCommand(), 0.6, false)
	return closest != "" && closest != missing
}

// gitBranchMissingNewCommand replaces the missing branch with the closest
// local branch.
func gitBranchMissingNewCommand(command *types.Command) ([]string, error) {
	if !gitBranchCommandUsesName(command.Script) {
		return nil, nil
	}
	missing, ok := gitBranchMissingName(command.Output)
	if !ok || missing == "" {
		return nil, nil
	}
	closest := utils.GetClosest(missing, gitLocalBranchesCommand(), 0.6, false)
	if closest == "" || closest == missing {
		return nil, nil
	}
	corrected := utils.ReplaceArgument(command.Script, missing, closest)
	if corrected == command.Script {
		return nil, nil
	}
	return []string{corrected}, nil
}

// gitBranchCommandUsesName reports whether script is a "git branch" subcommand
// that takes an existing branch name as its first argument.
func gitBranchCommandUsesName(script string) bool {
	for _, prefix := range []string{
		"branch -d ",
		"branch -D ",
		"branch --delete ",
		"branch -m ",
		"branch -M ",
		"branch --move ",
	} {
		if strings.Contains(script, prefix) {
			return true
		}
	}
	return false
}

// gitBranchMissingName returns the missing branch name from a git error.
func gitBranchMissingName(output string) (string, bool) {
	for _, pattern := range gitBranchMissingPatterns {
		if match := pattern.FindStringSubmatch(output); match != nil {
			return match[1], true
		}
	}
	return "", false
}

func init() {
	register(types.NewRule("git_branch_missing",
		utils.ForApp([]string{"git"}, 0, gitBranchMissingMatch),
		gitBranchMissingNewCommand))
}
