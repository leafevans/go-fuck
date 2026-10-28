package rules

import "testing"

// gitMergeUnrelatedOutput is the output git prints for unrelated histories.
func gitMergeUnrelatedOutput() string {
	return "fatal: refusing to merge unrelated histories"
}

func TestRuleGitMergeUnrelated(t *testing.T) {
	resetTestState(t)
	rule := Get("git_merge_unrelated")
	if rule == nil {
		t.Fatal("git_merge_unrelated rule is not registered")
	}

	assertMatch(t, rule, "git merge test", gitMergeUnrelatedOutput())
	assertNotMatch(t, rule, "git merge master", "")
	assertNotMatch(t, rule, "ls", gitMergeUnrelatedOutput())

	commandCases := []struct {
		script string
		want   string
	}{
		{"git merge local", "git merge local --allow-unrelated-histories"},
		{"git merge -m \"test\" local", "git merge -m \"test\" local --allow-unrelated-histories"},
		{"git merge -m \"test local\" local", "git merge -m \"test local\" local --allow-unrelated-histories"},
	}
	for _, testCase := range commandCases {
		t.Run("new_command/"+testCase.script, func(t *testing.T) {
			assertNewCommands(t, rule, testCase.script, gitMergeUnrelatedOutput(),
				[]string{testCase.want})
		})
	}
}
