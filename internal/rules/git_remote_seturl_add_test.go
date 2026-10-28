package rules

import "testing"

func TestRuleGitRemoteSeturlAdd(t *testing.T) {
	resetTestState(t)
	rule := Get("git_remote_seturl_add")
	if rule == nil {
		t.Fatal("git_remote_seturl_add rule is not registered")
	}

	t.Run("matches a missing remote", func(t *testing.T) {
		assertMatch(t, rule, "git remote set-url origin url", "fatal: No such remote")
	})

	t.Run("does not match other commands", func(t *testing.T) {
		cases := []struct {
			script string
			output string
		}{
			{"git remote set-url origin url", ""},
			{"git remote add origin url", ""},
			{"git remote remove origin", ""},
			{"git remote prune origin", ""},
			{"git remote set-branches origin branch", ""},
		}
		for _, testCase := range cases {
			assertNotMatch(t, rule, testCase.script, testCase.output)
		}
	})

	t.Run("new command adds the remote", func(t *testing.T) {
		assertNewCommands(t, rule,
			"git remote set-url origin git@github.com:nvbn/thefuck.git", "",
			[]string{"git remote add origin git@github.com:nvbn/thefuck.git"})
	})
}
