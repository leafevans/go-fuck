package rules

import "testing"

// gitTwoDashesOutput returns git's suggestion for a long option.
func gitTwoDashesOutput(option string) string {
	return "error: did you mean `" + option + "` (with two dashes ?)"
}

func TestRuleGitTwoDashes(t *testing.T) {
	resetTestState(t)
	rule := Get("git_two_dashes")
	if rule == nil {
		t.Fatal("git_two_dashes rule is not registered")
	}

	t.Run("matches a single dash option", func(t *testing.T) {
		output := gitTwoDashesOutput("--patch")
		assertMatch(t, rule, "git add -patch", output)
		assertMatch(t, rule, "git checkout -patch", output)
		assertMatch(t, rule, "git commit -amend", gitTwoDashesOutput("--amend"))
		assertMatch(t, rule, "git push -tags", gitTwoDashesOutput("--tags"))
		assertMatch(t, rule, "git rebase -continue", gitTwoDashesOutput("--continue"))
	})

	t.Run("current Git hint changes only the option", func(t *testing.T) {
		const output = "trace: inspecting `repo`\r\nerror: did you mean `--short` (with two dashes)?\r\n"
		const script = `git -C "my repo" status -short --branch -- -short`
		assertMatch(t, rule, script, output)
		assertNewCommands(t, rule, script, output, []string{`git -C "my repo" status --short --branch -- -short`})
	})

	t.Run("ignores hints without a matching option", func(t *testing.T) {
		for _, script := range []string{
			"git status --short",
			"git status -shorter",
			"git status -short -short",
			"git -C -short status",
			"git status -- -short",
			`git status -- "a -short path"`,
		} {
			output := gitTwoDashesOutput("--short")
			assertNotMatch(t, rule, script, output)
			assertNewCommands(t, rule, script, output, nil)
		}
	})

	t.Run("does not match long options", func(t *testing.T) {
		for _, script := range []string{
			"git add --patch",
			"git checkout --patch",
			"git commit --amend",
			"git push --tags",
			"git rebase --continue",
		} {
			assertNotMatch(t, rule, script, "")
		}
	})

	t.Run("new command uses the long option", func(t *testing.T) {
		cases := []struct {
			script string
			option string
			want   string
		}{
			{"git add -patch", "--patch", "git add --patch"},
			{"git checkout -patch", "--patch", "git checkout --patch"},
			{"git init -bare", "--bare", "git init --bare"},
			{"git commit -amend", "--amend", "git commit --amend"},
			{"git push -tags", "--tags", "git push --tags"},
			{"git rebase -continue", "--continue", "git rebase --continue"},
		}
		for _, testCase := range cases {
			assertNewCommands(t, rule, testCase.script, gitTwoDashesOutput(testCase.option),
				[]string{testCase.want})
		}
	})
}
