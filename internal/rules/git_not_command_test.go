package rules

import "testing"

// gitNotCommandOutput is the output git prints for an unknown command with a
// single suggestion.
func gitNotCommandOutput() string {
	return `git: 'brnch' is not a git command. See 'git --help'.

The most similar command is
branch
`
}

// gitNotCommandOneOfOutput is the output git prints for an unknown command
// with several suggestions.
func gitNotCommandOneOfOutput() string {
	return `git: 'st' is not a git command. See 'git --help'.

The most similar commands are
status
reset
stage
stash
stats
`
}

// gitNotCommandClosestOutput is the output git prints with tab indented
// suggestions.
func gitNotCommandClosestOutput() string {
	return "git: 'tags' is not a git command. See 'git --help'.\n\n" +
		"The most similar commands are\n" +
		"\tstage\n" +
		"\ttag\n"
}

func TestRuleGitNotCommand(t *testing.T) {
	resetTestState(t)
	rule := Get("git_not_command")
	if rule == nil {
		t.Fatal("git_not_command rule is not registered")
	}

	assertMatch(t, rule, "git brnch", gitNotCommandOutput())
	assertMatch(t, rule, "git st", gitNotCommandOneOfOutput())
	assertNotMatch(t, rule, "ls brnch", gitNotCommandOutput())
	assertNotMatch(t, rule, "git branch", "* master")

	commandCases := []struct {
		script string
		output string
		want   []string
	}{
		{"git brnch", gitNotCommandOutput(), []string{"git branch"}},
		{"git st", gitNotCommandOneOfOutput(),
			[]string{"git stats", "git stash", "git stage"}},
		{"git tags", gitNotCommandClosestOutput(),
			[]string{"git tag", "git stage"}},
	}
	for _, testCase := range commandCases {
		t.Run("new_command/"+testCase.script, func(t *testing.T) {
			assertNewCommands(t, rule, testCase.script, testCase.output, testCase.want)
		})
	}
}
