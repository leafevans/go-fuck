package rules

import "testing"

// gitPullCloneError is the output of git pull outside of a repository.
var gitPullCloneError = "\nfatal: Not a git repository (or any parent up to mount point /home)\n" +
	"Stopping at filesystem boundary (GIT_DISCOVERY_ACROSS_FILESYSTEM not set).\n"

func TestRuleGitPullClone(t *testing.T) {
	resetTestState(t)
	rule := Get("git_pull_clone")
	if rule == nil {
		t.Fatal("git_pull_clone rule is not registered")
	}

	t.Run("matches pull outside a repository", func(t *testing.T) {
		assertMatch(t, rule, "git pull git@github.com:mcarton/thefuck.git", gitPullCloneError)
	})

	t.Run("does not match without both markers", func(t *testing.T) {
		assertNotMatch(t, rule, "git pull git@github.com:mcarton/thefuck.git",
			"fatal: Not a git repository")
	})

	t.Run("does not match other applications", func(t *testing.T) {
		assertNotMatch(t, rule, "ls", gitPullCloneError)
	})

	t.Run("new command clones the repository", func(t *testing.T) {
		assertNewCommands(t, rule, "git pull git@github.com:mcarton/thefuck.git",
			gitPullCloneError,
			[]string{"git clone git@github.com:mcarton/thefuck.git"})
	})
}
