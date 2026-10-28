package rules

import "testing"

// gitMergeOutput is the output git prints for an unknown merge branch.
func gitMergeOutput() string {
	return "merge: local - not something we can merge\n\n" +
		"Did you mean this?\n\tremote/local"
}

func TestRuleGitMerge(t *testing.T) {
	resetTestState(t)
	rule := Get("git_merge")
	if rule == nil {
		t.Fatal("git_merge rule is not registered")
	}

	assertMatch(t, rule, "git merge test", gitMergeOutput())
	assertNotMatch(t, rule, "git merge master", "")
	assertNotMatch(t, rule, "ls", gitMergeOutput())

	commandCases := []struct {
		script string
		want   string
	}{
		{"git merge local", "git merge remote/local"},
		{"git merge -m \"test\" local", "git merge -m \"test\" remote/local"},
		{"git merge -m \"test local\" local", "git merge -m \"test local\" remote/local"},
	}
	for _, testCase := range commandCases {
		t.Run("new_command/"+testCase.script, func(t *testing.T) {
			assertNewCommands(t, rule, testCase.script, gitMergeOutput(),
				[]string{testCase.want})
		})
	}
}
