package rules

import "testing"

func TestRuleGitHookBypass(t *testing.T) {
	resetTestState(t)
	rule := Get("git_hook_bypass")
	if rule == nil {
		t.Fatal("git_hook_bypass rule is not registered")
	}
	if rule.Priority != 1100 {
		t.Errorf("git_hook_bypass priority = %d, want 1100", rule.Priority)
	}
	if !rule.RequiresOutput {
		t.Error("git_hook_bypass must require the command output")
	}

	hookOutput := "husky - pre-commit hook exited with code 1 (error)"
	for _, script := range []string{
		"git am",
		"git commit",
		"git commit -m 'foo bar'",
		"git push",
		"git push -u foo bar",
	} {
		t.Run("match/"+script, func(t *testing.T) {
			assertMatch(t, rule, script, hookOutput)
		})
	}

	for _, script := range []string{"git add foo", "git status", "git diff foo bar"} {
		t.Run("not_match/"+script, func(t *testing.T) {
			assertNotMatch(t, rule, script, hookOutput)
		})
	}
	// A failure that a hook did not cause must not suggest bypassing hooks.
	assertNotMatch(t, rule, "git push", "! [rejected]        main -> main (fetch first)\nfatal: failed to push some refs to 'origin'")
	assertNotMatch(t, rule, "git commit -m x", "Author identity unknown")
	assertNotMatch(t, rule, "git commit -m x", "")
	// Mentioning a hook or tool is not evidence that a bypassable hook failed.
	assertNotMatch(t, rule, "git push", "fatal: repository hook-demo not found")
	assertNotMatch(t, rule, "git push", "remote: pre-receive hook declined")
	assertNotMatch(t, rule, "git commit", "husky installed successfully")
	assertNotMatch(t, rule, "git push", "error: failed to push branch lint-staged")
	// The flag is already there.
	assertNotMatch(t, rule, "git commit --no-verify -m x", hookOutput)
	assertNotMatch(t, rule, "git log commit", "")
	assertNewCommands(t, rule, "git -C 'repo dir' commit -m 'commit message'", "", []string{"git -C 'repo dir' commit --no-verify -m 'commit message'"})

	commandCases := []struct {
		script string
		want   string
	}{
		{"git am", "git am --no-verify"},
		{"git commit", "git commit --no-verify"},
		{"git commit -m 'foo bar'", "git commit --no-verify -m 'foo bar'"},
		{"git push", "git push --no-verify"},
		{"git push -p", "git push --no-verify -p"},
	}
	for _, testCase := range commandCases {
		t.Run("new_command/"+testCase.script, func(t *testing.T) {
			assertNewCommands(t, rule, testCase.script, "", []string{testCase.want})
		})
	}
}
