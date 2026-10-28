package rules

import (
	"fmt"
	"testing"
)

// gitPushDifferentBranchOutput returns git's output when the upstream branch
// has a different name than the current branch.
func gitPushDifferentBranchOutput(localBranch, remoteBranch string) string {
	return fmt.Sprintf("fatal: The upstream branch of your current branch does not match\n"+
		"the name of your current branch.  To push to the upstream branch\n"+
		"on the remote, use\n\n"+
		"    git push origin HEAD:%s\n\n"+
		"To push to the branch of the same name on the remote, use\n\n"+
		"    git push origin %s\n\n"+
		"To choose either option permanently, see push.default in 'git help config'.\n",
		remoteBranch, localBranch)
}

func TestRuleGitPushDifferentBranchNames(t *testing.T) {
	resetTestState(t)
	rule := Get("git_push_different_branch_names")
	if rule == nil {
		t.Fatal("git_push_different_branch_names rule is not registered")
	}

	t.Run("matches different branch names", func(t *testing.T) {
		assertMatch(t, rule, "git push", gitPushDifferentBranchOutput("foo", "bar"))
	})

	t.Run("does not match other commands", func(t *testing.T) {
		output := gitPushDifferentBranchOutput("foo", "bar")
		assertNotMatch(t, rule, "vim", "")
		assertNotMatch(t, rule, "git status", output)
		assertNotMatch(t, rule, "git push", "")
	})

	t.Run("new command uses the first suggestion", func(t *testing.T) {
		assertNewCommands(t, rule, "git push", gitPushDifferentBranchOutput("foo", "bar"),
			[]string{"git push origin HEAD:bar"})
	})
}
