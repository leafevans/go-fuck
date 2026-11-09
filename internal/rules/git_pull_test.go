package rules

import (
	"strings"
	"testing"
)

// gitPullOutput returns the output of git pull on a branch without upstream.
func gitPullOutput() string {
	return "There is no tracking information for the current branch.\n" +
		"Please specify which branch you want to merge with.\n" +
		"See git-pull(1) for details\n\n" +
		"    git pull <remote> <branch>\n\n" +
		"If you wish to set tracking information for this branch you can do so with:\n\n" +
		"    git branch --set-upstream-to=<remote>/<branch> master\n\n"
}

func TestRuleGitPull(t *testing.T) {
	resetTestState(t)
	rule := Get("git_pull")
	if rule == nil {
		t.Fatal("git_pull rule is not registered")
	}

	t.Run("matches missing upstream", func(t *testing.T) {
		assertMatch(t, rule, "git pull", gitPullOutput())
	})

	t.Run("does not match without output", func(t *testing.T) {
		assertNotMatch(t, rule, "git pull", "")
	})

	t.Run("does not match other applications", func(t *testing.T) {
		assertNotMatch(t, rule, "ls", gitPullOutput())
	})

	t.Run("new command sets the upstream branch", func(t *testing.T) {
		assertNewCommands(t, rule, "git pull", gitPullOutput(),
			[]string{"git branch --set-upstream-to=origin/master master && git pull"})
	})
}

// TestRuleGitPullToleratesTrailingLines guards that the suggested upstream
// command is found by its prefix, because a re-run through PowerShell may
// append empty lines to the captured output.
func TestRuleGitPullToleratesTrailingLines(t *testing.T) {
	resetTestState(t)
	rule := Get("git_pull")
	if rule == nil {
		t.Fatal("git_pull rule is not registered")
	}

	want := []string{"git branch --set-upstream-to=origin/master master && git pull"}
	for _, output := range []string{
		gitPullOutput() + "\n\n\n",
		strings.ReplaceAll(gitPullOutput(), "\n", "\r\n"),
		strings.TrimRight(gitPullOutput(), "\n") + "\n",
	} {
		assertNewCommands(t, rule, "git pull", output, want)
	}
}
