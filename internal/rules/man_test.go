package rules

import "testing"

func TestRuleMan(t *testing.T) {
	resetTestState(t)
	rule := Get("man")
	if rule == nil {
		t.Fatal("man rule is not registered")
	}

	for _, script := range []string{
		"man read",
		"man 2 read",
		"man 3 read",
		"man -s2 read",
		"man -s3 read",
		"man -s 2 read",
		"man -s 3 read",
	} {
		assertMatch(t, rule, script, "")
	}
	for _, script := range []string{"man", "man "} {
		assertNotMatch(t, rule, script, "")
	}

	for _, testCase := range []struct {
		script string
		output string
		want   []string
	}{
		{"man read", "", []string{"man 3 read", "man 2 read", "read --help"}},
		{"man missing", "No manual entry for missing\n", []string{"missing --help"}},
		{"man 2 read", "", []string{"man 3 read"}},
		{"man 3 read", "", []string{"man 2 read"}},
		{"man -s2 read", "", []string{"man -s3 read"}},
		{"man -s3 read", "", []string{"man -s2 read"}},
		{"man -s 2 read", "", []string{"man -s 3 read"}},
		{"man -s 3 read", "", []string{"man -s 2 read"}},
	} {
		assertNewCommands(t, rule, testCase.script, testCase.output, testCase.want)
	}
}
