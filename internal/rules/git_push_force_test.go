package rules

import "testing"

// gitPushForceError is the output of a rejected non-fast-forward push.
var gitPushForceError = "\nTo /tmp/foo\n" +
	" ! [rejected]        master -> master (non-fast-forward)\n" +
	" error: failed to push some refs to '/tmp/bar'\n" +
	" hint: Updates were rejected because the tip of your current branch is behind\n" +
	" hint: its remote counterpart. Integrate the remote changes (e.g.\n" +
	" hint: 'git pull ...') before pushing again.\n" +
	" hint: See the 'Note about fast-forwards' in 'git push --help' for details.\n"

// gitPushForceOK is the output of a successful push.
var gitPushForceOK = "\nCounting objects: 3, done.\n" +
	"Delta compression using up to 4 threads.\n" +
	"Compressing objects: 100% (2/2), done.\n" +
	"Writing objects: 100% (3/3), 282 bytes | 0 bytes/s, done.\n" +
	"Total 3 (delta 0), reused 0 (delta 0)\n" +
	"To /tmp/bar\n" +
	"   514eed3..f269c79  master -> master\n"

func TestRuleGitPushForce(t *testing.T) {
	resetTestState(t)
	rule := Get("git_push_force")
	if rule == nil {
		t.Fatal("git_push_force rule is not registered")
	}

	t.Run("matches a rejected push", func(t *testing.T) {
		assertMatch(t, rule, "git push", gitPushForceError)
		assertMatch(t, rule, "git push nvbn", gitPushForceError)
		assertMatch(t, rule, "git push nvbn master", gitPushForceError)
	})

	t.Run("does not match successful pushes", func(t *testing.T) {
		assertNotMatch(t, rule, "git push", gitPushForceOK)
		assertNotMatch(t, rule, "git push", "Everything up-to-date")
		assertNotMatch(t, rule, "git push nvbn", gitPushForceOK)
		assertNotMatch(t, rule, "git push nvbn master", "Everything up-to-date")
	})

	t.Run("new command adds --force-with-lease", func(t *testing.T) {
		cases := []struct {
			script string
			want   string
		}{
			{"git push", "git push --force-with-lease"},
			{"git push nvbn", "git push --force-with-lease nvbn"},
			{"git push nvbn master", "git push --force-with-lease nvbn master"},
		}
		for _, testCase := range cases {
			assertNewCommands(t, rule, testCase.script, gitPushForceError,
				[]string{testCase.want})
		}
	})
}
