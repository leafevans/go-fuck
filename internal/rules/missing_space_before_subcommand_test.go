package rules

import "testing"

// missingSpaceFakeExecutables returns the installed executables used by the
// tests.
func missingSpaceFakeExecutables() []string {
	return []string{"git", "ls", "npm", "w", "watch"}
}

func TestRuleMissingSpaceBeforeSubcommand(t *testing.T) {
	resetTestState(t)
	rule := Get("missing_space_before_subcommand")
	if rule == nil {
		t.Fatal("missing_space_before_subcommand rule is not registered")
	}

	original := missingSpaceGetAllExecutables
	missingSpaceGetAllExecutables = missingSpaceFakeExecutables
	t.Cleanup(func() { missingSpaceGetAllExecutables = original })

	for _, script := range []string{"gitbranch", "ls-la", "npminstall", "watchls"} {
		assertMatch(t, rule, script, "")
	}
	for _, script := range []string{"git branch", "vimfile"} {
		assertNotMatch(t, rule, script, "")
	}

	for _, testCase := range []struct {
		script string
		want   string
	}{
		{"gitbranch", "git branch"},
		{"ls-la", "ls -la"},
		{"npminstall webpack", "npm install webpack"},
		{"watchls", "watch ls"},
	} {
		assertNewCommands(t, rule, testCase.script, "", []string{testCase.want})
	}
}
