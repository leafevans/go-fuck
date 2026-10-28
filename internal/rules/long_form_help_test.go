package rules

import "testing"

func TestRuleLongFormHelp(t *testing.T) {
	resetTestState(t)
	rule := Get("long_form_help")
	if rule == nil {
		t.Fatal("long_form_help rule is not registered")
	}

	assertMatch(t, rule, "grep -h", "Try 'grep --help' for more information.")
	assertNotMatch(t, rule, "", "")

	for _, testCase := range []struct {
		before string
		after  string
	}{
		{"grep -h", "grep --help"},
		{"tar -h", "tar --help"},
		{"docker run -h", "docker run --help"},
		{"cut -h", "cut --help"},
	} {
		assertNewCommands(t, rule, testCase.before, "", []string{testCase.after})
	}
}
