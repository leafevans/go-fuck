package rules

import "testing"

func TestRuleGitPushWithoutCommits(t *testing.T) {
	resetTestState(t)
	rule := Get("git_push_without_commits")
	if rule == nil {
		t.Fatal("git_push_without_commits rule is not registered")
	}

	const (
		script = "git push -u origin master"
		output = "error: src refspec master does not match any\nerror: failed to..."
	)

	t.Run("matches missing commits", func(t *testing.T) {
		assertMatch(t, rule, script, output)
	})

	t.Run("does not match an up-to-date push", func(t *testing.T) {
		assertNotMatch(t, rule, script, "Everything up-to-date")
	})

	t.Run("new command creates the initial commit", func(t *testing.T) {
		assertNewCommands(t, rule, script, output,
			[]string{"git commit -m \"Initial commit\" && " + script})
	})
}
