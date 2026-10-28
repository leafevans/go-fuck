package rules

import "testing"

func TestRuleYarnAlias(t *testing.T) {
	resetTestState(t)

	rule := Get("yarn_alias")
	if rule == nil {
		t.Fatal("yarn_alias rule is not registered")
	}

	cases := []struct {
		script string
		output string
		want   string
	}{
		{"yarn rm", "error Did you mean `yarn remove`?", "yarn remove"},
		{"yarn etil", "error Command \"etil\" not found. Did you mean \"etl\"?", "yarn etl"},
		{"yarn ls", "error Did you mean `yarn list`?", "yarn list"},
	}
	for _, testCase := range cases {
		assertMatch(t, rule, testCase.script, testCase.output)
		assertNewCommands(t, rule, testCase.script, testCase.output,
			[]string{testCase.want})
	}

	assertNotMatch(t, rule, "yarn rm", "")
	assertNotMatch(t, rule, "npm rm", "error Did you mean `yarn remove`?")
}
