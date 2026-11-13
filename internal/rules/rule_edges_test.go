package rules

import "testing"

func TestRuleEdgeCases(t *testing.T) {
	resetTestState(t)

	notMatching := []struct {
		rule   string
		script string
		output string
	}{
		{"git_push", "git status", "git push --set-upstream origin main"},
		{"apt_upgrade", "apt list", "Listing...\nfoo/stable 1.0 amd64"},
		{"git_worktree_remove_missing", "git status", ""},
		{"git_worktree_branch_used", "git status", ""},
		{"grep_recursive", "grep", ""},
		{"just_unknown_recipe", "just task", ""},
	}
	for _, testCase := range notMatching {
		rule := Get(testCase.rule)
		if rule == nil {
			t.Fatalf("%s rule is not registered", testCase.rule)
		}
		if rule.Match(testCommand(testCase.script, testCase.output)) {
			t.Errorf("%s should not match %q", testCase.rule, testCase.script)
		}
	}

	nilCommands := []struct {
		rule   string
		script string
		output string
	}{
		{"git_worktree_remove_missing", "git status", ""},
		{"git_worktree_branch_used", "git status", ""},
		{"powershell_mkdir_existing", "", ""},
		{"grep_recursive", "grep", ""},
		{"no_such_file", "cat missing", ""},
		{"just_unknown_recipe", "just task", ""},
		{"git_branch_missing", "git status", ""},
		{"git_tag_missing", "git status", ""},
		{"git_unknown_remote", "git status", ""},
		{"task_unknown_task", "task", ""},
		{"yarn_alias", "yarn", ""},
		{"npm_wrong_command", "", ""},
		{"gradle_no_task", "gradle", ""},
		{"missing_space_before_subcommand", "", ""},
	}
	for _, testCase := range nilCommands {
		rule := Get(testCase.rule)
		got, err := rule.GetNewCommand(testCommand(testCase.script, testCase.output))
		if err != nil {
			t.Fatalf("%s returned an error: %v", testCase.rule, err)
		}
		if len(got) != 0 {
			t.Errorf("%s returned %q, want no commands", testCase.rule, got)
		}
	}
}

func TestHostscliNewCommandEdges(t *testing.T) {
	resetTestState(t)
	rule := Get("hostscli")
	if rule == nil {
		t.Fatal("hostscli rule is not registered")
	}
	if got, _ := rule.GetNewCommand(testCommand("hostscli websites", hostscliNoWebsite)); len(got) != 1 ||
		got[0] != "hostscli websites" {
		t.Fatalf("hostscliNewCommand(website) = %q", got)
	}
	if got, _ := rule.GetNewCommand(testCommand("hostscli", "other error")); len(got) != 0 {
		t.Fatalf("hostscliNewCommand(other) = %q", got)
	}
	if got, _ := rule.GetNewCommand(testCommand("hostscli blok", "Error: No such command \"blok\"")); len(got) == 0 {
		t.Fatal("hostscliNewCommand(misspelled) returned no suggestions")
	}
}

func TestLongFormHelpNewCommandEdges(t *testing.T) {
	resetTestState(t)
	rule := Get("long_form_help")
	if rule == nil {
		t.Fatal("long_form_help rule is not registered")
	}
	if got, _ := rule.GetNewCommand(testCommand("tool", "Run 'tool --help' for details.")); len(got) != 1 ||
		got[0] != "tool --help" {
		t.Fatalf("longFormHelpNewCommand(suggestion) = %q", got)
	}
	if got, _ := rule.GetNewCommand(testCommand("tool", "RUN 'tool --help' for details.")); len(got) != 0 {
		t.Fatalf("longFormHelpNewCommand(uppercase suggestion) = %q", got)
	}
	if got, _ := rule.GetNewCommand(testCommand("tool -h", "see --help")); len(got) != 1 ||
		got[0] != "tool --help" {
		t.Fatalf("longFormHelpNewCommand(short flag) = %q", got)
	}
	if got, _ := rule.GetNewCommand(testCommand("tool", "see --help")); len(got) != 0 {
		t.Fatalf("longFormHelpNewCommand(no short flag) = %q", got)
	}
}

func TestDirtyUnzipZipFileEdges(t *testing.T) {
	resetTestState(t)
	cases := []struct {
		script string
		want   string
	}{
		{"unzip", ""},
		{"unzip -q", ""},
		{"unzip -q archive", "archive.zip"},
		{"unzip archive.zip -d dir", "archive.zip"},
	}
	for _, testCase := range cases {
		if got := dirtyUnzipZipFile(testCommand(testCase.script, "")); got != testCase.want {
			t.Errorf("dirtyUnzipZipFile(%q) = %q, want %q", testCase.script, got, testCase.want)
		}
	}
}
