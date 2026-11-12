package rules

import (
	"regexp"
	"strings"

	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// gitWorktreeBranchUsedPattern matches the error that git prints when a branch
// is already checked out in another working tree.
var gitWorktreeBranchUsedPattern = regexp.MustCompile("(?m)^fatal: '([^']+)' is already used by worktree at '([^']+)'$")

// gitWorktreeBranchUsedMatch reports whether "git worktree add" failed because
// the branch is already checked out in another worktree.
func gitWorktreeBranchUsedMatch(command *types.Command) bool {
	return strings.Contains(command.Script, "worktree add ") &&
		gitWorktreeBranchUsedPattern.MatchString(command.Output)
}

// gitWorktreeBranchUsedNewCommand forces the new working tree. The user asked
// for this branch explicitly, so reusing it in a second worktree is the
// intended operation.
func gitWorktreeBranchUsedNewCommand(command *types.Command) ([]string, error) {
	if !gitWorktreeBranchUsedMatch(command) {
		return nil, nil
	}
	corrected := strings.Replace(command.Script, "worktree add ", "worktree add --force ", 1)
	if corrected == command.Script {
		return nil, nil
	}
	return []string{corrected}, nil
}

func init() {
	register(types.NewRule("git_worktree_branch_used",
		utils.ForApp([]string{"git"}, 0, gitWorktreeBranchUsedMatch),
		gitWorktreeBranchUsedNewCommand))
}
