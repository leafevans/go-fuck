package rules

import "testing"

func TestRuleGitAdd(t *testing.T) {
	resetTestState(t)
	rule := Get("git_add")
	if rule == nil {
		t.Fatal("git_add rule is not registered")
	}

	original := gitAddFileExists
	t.Cleanup(func() { gitAddFileExists = original })
	gitAddFileExists = func(string) bool { return true }

	output := func(target string) string {
		return "error: pathspec '" + target + "' did not match any file(s) known to git."
	}

	for _, script := range []string{"git submodule update unknown", "git commit unknown"} {
		t.Run("match/"+script, func(t *testing.T) {
			assertMatch(t, rule, script, output("unknown"))
		})
	}

	// An empty pathspec exists on disk but must not be suggested.
	for _, script := range []string{"git submodule update known", "git commit known"} {
		t.Run("not_match_empty_pathspec/"+script, func(t *testing.T) {
			assertNotMatch(t, rule, script, output(""))
		})
	}

	gitAddFileExists = func(string) bool { return false }
	for _, script := range []string{"git submodule update known", "git commit known"} {
		t.Run("not_match_missing_file/"+script, func(t *testing.T) {
			assertNotMatch(t, rule, script, output("known"))
		})
	}
	gitAddFileExists = func(string) bool { return true }

	assertNotMatch(t, rule, "git commit unknown", "")

	commandCases := []struct {
		script string
		want   string
	}{
		{"git submodule update unknown", "git add -- unknown && git submodule update unknown"},
		{"git commit unknown", "git add -- unknown && git commit unknown"},
	}
	for _, testCase := range commandCases {
		t.Run("new_command/"+testCase.script, func(t *testing.T) {
			assertNewCommands(t, rule, testCase.script, output("unknown"),
				[]string{testCase.want})
		})
	}
}
