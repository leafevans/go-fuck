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
			"git log --name-only -p README.md"},
		{"git log README.md -p", "fatal: option '-p' must come before non-option arguments",
			"git log -p README.md"},
		{"git log README.md -p CONTRIBUTING.md", "fatal: option '-p' must come before non-option arguments",
			"git log -p README.md CONTRIBUTING.md"},
		{"git log -p README.md --name-only", "fatal: option '--name-only' must come before non-option arguments",
			"git log --name-only -p README.md"},
	}
	assertNewCommands(t, rule, "git -C 'repo dir' log 'file name' -p other", "fatal: bad flag '-p' used after filename", []string{"git -C 'repo dir' log -p 'file name' other"})
	assertNotMatch(t, rule, "git log README.md --format '%s'", "fatal: option '--format' must come before non-option arguments")
	assertNotMatch(t, rule, "git log -p file", "fatal: bad flag '-p' used after filename")
	assertNotMatch(t, rule, "git log -- file -p", "fatal: bad flag '-p' used after filename")
	assertNewCommands(t, rule, "git log 'the file'\t\t--format=%s", "fatal: option '--format=%s' must come before non-option arguments", []string{"git log --format=%s 'the file'"})
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
