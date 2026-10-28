package rules

import "testing"

func TestRuleWrongHyphenBeforeSubcommand(t *testing.T) {
	resetTestState(t)
	rule := Get("wrong_hyphen_before_subcommand")
	if rule == nil {
		t.Fatal("wrong_hyphen_before_subcommand rule is not registered")
	}
	if rule.RequiresOutput {
		t.Error("wrong_hyphen_before_subcommand rule must not require the command output")
	}
	if rule.Priority != 4500 {
		t.Errorf("wrong_hyphen_before_subcommand priority is %d, want 4500", rule.Priority)
	}

	original := wrongHyphenGetAllExecutables
	wrongHyphenGetAllExecutables = func() []string {
		return []string{"git", "apt", "apt-get", "ls", "pwd"}
	}
	t.Cleanup(func() { wrongHyphenGetAllExecutables = original })

	for _, script := range []string{"git-log", "apt-install python"} {
		assertMatch(t, rule, script, "")
	}
	for _, script := range []string{"ls -la", "git2-make", "apt-get install python"} {
		assertNotMatch(t, rule, script, "")
	}

	for _, testCase := range []struct {
		script string
		want   string
	}{
		{"git-log", "git log"},
		{"apt-install python", "apt install python"},
	} {
		assertNewCommands(t, rule, testCase.script, "", []string{testCase.want})
	}
}
