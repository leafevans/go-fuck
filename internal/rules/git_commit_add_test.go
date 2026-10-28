package rules

import "testing"

func TestRuleGitCommitAdd(t *testing.T) {
	resetTestState(t)
	rule := Get("git_commit_add")
	if rule == nil {
		t.Fatal("git_commit_add rule is not registered")
	}

	matchCases := []struct {
		script string
		output string
	}{
		{"git commit -m \"test\"", "no changes added to commit"},
		{"git commit", "no changes added to commit"},
	}
	for _, testCase := range matchCases {
		t.Run("match/"+testCase.script, func(t *testing.T) {
			assertMatch(t, rule, testCase.script, testCase.output)
		})
	}

	notMatchCases := []struct {
		script string
		output string
	}{
		{"git commit -m \"test\"", " 1 file changed, 15 insertions(+), 14 deletions(-)"},
		{"git branch foo", ""},
		{"git checkout feature/test_commit", ""},
		{"git push", ""},
	}
	for _, testCase := range notMatchCases {
		t.Run("not_match/"+testCase.script, func(t *testing.T) {
			assertNotMatch(t, rule, testCase.script, testCase.output)
		})
	}

	commandCases := []struct {
		script string
		want   []string
	}{
		{"git commit", []string{"git commit -a", "git commit -p"}},
		{"git commit -m \"foo\"", []string{
			"git commit -a -m \"foo\"", "git commit -p -m \"foo\""}},
	}
	for _, testCase := range commandCases {
		t.Run("new_command/"+testCase.script, func(t *testing.T) {
			assertNewCommands(t, rule, testCase.script, "", testCase.want)
		})
	}
}
