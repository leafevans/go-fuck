package rules

import "testing"

func TestRuleGitFlagAfterFilename(t *testing.T) {
	resetTestState(t)
	rule := Get("git_flag_after_filename")
	if rule == nil {
		t.Fatal("git_flag_after_filename rule is not registered")
	}

	cases := []struct {
		script string
		output string
		want   string
	}{
		{"git log README.md -p", "fatal: bad flag '-p' used after filename",
			"git log -p README.md"},
		{"git log README.md -p CONTRIBUTING.md", "fatal: bad flag '-p' used after filename",
			"git log -p README.md CONTRIBUTING.md"},
		{"git log -p README.md --name-only", "fatal: bad flag '--name-only' used after filename",
			"git log -p --name-only README.md"},
		{"git log README.md -p", "fatal: option '-p' must come before non-option arguments",
			"git log -p README.md"},
		{"git log README.md -p CONTRIBUTING.md", "fatal: option '-p' must come before non-option arguments",
			"git log -p README.md CONTRIBUTING.md"},
		{"git log -p README.md --name-only", "fatal: option '--name-only' must come before non-option arguments",
			"git log -p --name-only README.md"},
	}
	for _, testCase := range cases {
		t.Run("match/"+testCase.script, func(t *testing.T) {
			assertMatch(t, rule, testCase.script, testCase.output)
		})
		t.Run("new_command/"+testCase.script, func(t *testing.T) {
			assertNewCommands(t, rule, testCase.script, testCase.output,
				[]string{testCase.want})
		})
	}

	for _, script := range []string{"git log README.md", "git log -p README.md"} {
		t.Run("not_match/"+script, func(t *testing.T) {
			assertNotMatch(t, rule, script, "")
		})
	}
}
