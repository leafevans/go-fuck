package rules

import "testing"

func TestRuleRemoveShellPromptLiteral(t *testing.T) {
	resetTestState(t)
	rule := Get("remove_shell_prompt_literal")
	if rule == nil {
		t.Fatal("remove_shell_prompt_literal rule is not registered")
	}

	const output = "$: command not found"

	for _, script := range []string{"$ cd newdir", " $ cd newdir", "$ $ cd newdir", " $ $ cd newdir"} {
		assertMatch(t, rule, script, output)
	}
	for _, testCase := range []struct {
		script string
		output string
	}{
		{"$", output},
		{" $", output},
		{"$?", "127: command not found"},
		{" $?", "127: command not found"},
		{"", ""},
	} {
		assertNotMatch(t, rule, testCase.script, testCase.output)
	}

	for _, testCase := range []struct {
		script string
		want   string
	}{
		{"$ cd newdir", "cd newdir"},
		{"$ $ cd newdir", "cd newdir"},
		{"$ python3 -m virtualenv env", "python3 -m virtualenv env"},
		{" $ $ $ python3 -m virtualenv env", "python3 -m virtualenv env"},
	} {
		assertNewCommands(t, rule, testCase.script, output, []string{testCase.want})
	}
}
