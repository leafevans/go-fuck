package rules

import "testing"

// gitRebaseMergeDirOutput is the output of git rebase while another rebase is
// in progress.
func gitRebaseMergeDirOutput() string {
	return "\n\nIt seems that there is already a rebase-merge directory, and\n" +
		"I wonder if you are in the middle of another rebase.  If that is the\n" +
		"case, please try\n" +
		"\tgit rebase (--continue | --abort | --skip)\n" +
		"If that is not the case, please\n" +
		"\trm -fr \"/foo/bar/baz/egg/.git/rebase-merge\"\n" +
		"and run me again.  I am stopping in case you still have something\n" +
		"valuable there.\n"
}

func TestRuleGitRebaseMergeDir(t *testing.T) {
	resetTestState(t)
	rule := Get("git_rebase_merge_dir")
	if rule == nil {
		t.Fatal("git_rebase_merge_dir rule is not registered")
	}

	output := gitRebaseMergeDirOutput()

	t.Run("matches a pending rebase", func(t *testing.T) {
		for _, script := range []string{
			"git rebase master",
			"git rebase -skip",
			"git rebase",
		} {
			assertMatch(t, rule, script, output)
		}
	})

	t.Run("does not match without output", func(t *testing.T) {
		assertNotMatch(t, rule, "git rebase master", "")
		assertNotMatch(t, rule, "git rebase -abort", "")
	})

	t.Run("new command orders suggestions by similarity", func(t *testing.T) {
		cases := []struct {
			script string
			want   []string
		}{
			{"git rebase master", []string{
				"git rebase --abort",
				"git rebase --skip",
				"git rebase --continue",
				"rm -fr \"/foo/bar/baz/egg/.git/rebase-merge\"",
			}},
			{"git rebase -skip", []string{
				"git rebase --skip",
				"git rebase --abort",
				"git rebase --continue",
				"rm -fr \"/foo/bar/baz/egg/.git/rebase-merge\"",
			}},
			{"git rebase", []string{
				"git rebase --skip",
				"git rebase --abort",
				"git rebase --continue",
				"rm -fr \"/foo/bar/baz/egg/.git/rebase-merge\"",
			}},
		}
		for _, testCase := range cases {
			assertNewCommands(t, rule, testCase.script, output, testCase.want)
		}
	})
}
