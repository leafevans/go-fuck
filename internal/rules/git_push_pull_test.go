package rules

import "testing"

// gitPushPullError is the output of a push rejected because the local branch
// is behind the remote one.
var gitPushPullError = "\nTo /tmp/foo\n" +
	" ! [rejected]        master -> master (non-fast-forward)\n" +
	" error: failed to push some refs to '/tmp/bar'\n" +
	" hint: Updates were rejected because the tip of your current branch is behind\n" +
	" hint: its remote counterpart. Integrate the remote changes (e.g.\n" +
	" hint: 'git pull ...') before pushing again.\n" +
	" hint: See the 'Note about fast-forwards' in 'git push --help' for details.\n"

// gitPushPullRemoteWorkError is the output of a push rejected because the
// remote contains work that is missing locally.
var gitPushPullRemoteWorkError = "\nTo /tmp/foo\n" +
	" ! [rejected]        master -> master (non-fast-forward)\n" +
	" error: failed to push some refs to '/tmp/bar'\n" +
	"hint: Updates were rejected because the remote contains work that you do\n" +
	"hint: not have locally. This is usually caused by another repository pushing\n" +
	"hint: to the same ref. You may want to first integrate the remote changes\n" +
	"hint: (e.g., 'git pull ...') before pushing again.\n" +
	"hint: See the 'Note about fast-forwards' in 'git push --help' for details.\n"

// gitPushPullOK is the output of a successful push.
var gitPushPullOK = "\nCounting objects: 3, done.\n" +
	"Delta compression using up to 4 threads.\n" +
	"Compressing objects: 100% (2/2), done.\n" +
	"Writing objects: 100% (3/3), 282 bytes | 0 bytes/s, done.\n" +
	"Total 3 (delta 0), reused 0 (delta 0)\n" +
	"To /tmp/bar\n" +
	"   514eed3..f269c79  master -> master\n"

func TestRuleGitPushPull(t *testing.T) {
	resetTestState(t)
	rule := Get("git_push_pull")
	if rule == nil {
		t.Fatal("git_push_pull rule is not registered")
	}

	t.Run("matches rejected pushes", func(t *testing.T) {
		for _, output := range []string{gitPushPullError, gitPushPullRemoteWorkError} {
			assertMatch(t, rule, "git push", output)
			assertMatch(t, rule, "git push nvbn", output)
			assertMatch(t, rule, "git push nvbn master", output)
		}
	})

	t.Run("does not match successful pushes", func(t *testing.T) {
		assertNotMatch(t, rule, "git push", gitPushPullOK)
		assertNotMatch(t, rule, "git push", "Everything up-to-date")
		assertNotMatch(t, rule, "git push nvbn", gitPushPullOK)
		assertNotMatch(t, rule, "git push nvbn master", "Everything up-to-date")
	})

	t.Run("new command pulls before pushing", func(t *testing.T) {
		cases := []struct {
			script string
			want   string
		}{
			{"git push", "git pull && git push"},
			{"git push nvbn", "git pull nvbn && git push nvbn"},
			{"git push nvbn master", "git pull nvbn master && git push nvbn master"},
		}
		for _, output := range []string{gitPushPullError, gitPushPullRemoteWorkError} {
			for _, testCase := range cases {
				assertNewCommands(t, rule, testCase.script, output,
					[]string{testCase.want})
			}
		}
	})
}
