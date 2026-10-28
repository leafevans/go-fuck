package rules

import "testing"

func TestRuleDry(t *testing.T) {
	resetTestState(t)
	rule := Get("dry")
	if rule == nil {
		t.Fatal("dry rule is not registered")
	}

	for _, script := range []string{"cd cd foo", "git git push origin/master"} {
		assertMatch(t, rule, script, "")
	}
	for _, script := range []string{"cd foo", "git push origin/master", ""} {
		assertNotMatch(t, rule, script, "")
	}

	for _, testCase := range []struct {
		script string
		want   string
	}{
		{"cd cd foo", "cd foo"},
		{"git git push origin/master", "git push origin/master"},
	} {
		assertNewCommands(t, rule, testCase.script, "", []string{testCase.want})
	}
}
