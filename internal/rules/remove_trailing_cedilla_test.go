package rules

import "testing"

func TestRuleRemoveTrailingCedilla(t *testing.T) {
	resetTestState(t)
	rule := Get("remove_trailing_cedilla")
	if rule == nil {
		t.Fatal("remove_trailing_cedilla rule is not registered")
	}

	for _, script := range []string{"wrong" + removeTrailingCedillaCedilla, "wrong with args" + removeTrailingCedillaCedilla} {
		assertMatch(t, rule, script, "")
	}
	for _, script := range []string{"wrong", "wrong with args", ""} {
		assertNotMatch(t, rule, script, "")
	}

	for _, testCase := range []struct {
		script string
		want   string
	}{
		{"wrong" + removeTrailingCedillaCedilla, "wrong"},
		{"wrong with args" + removeTrailingCedillaCedilla, "wrong with args"},
	} {
		assertNewCommands(t, rule, testCase.script, "", []string{testCase.want})
	}
}
