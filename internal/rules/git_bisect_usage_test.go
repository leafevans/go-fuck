package rules

import "testing"

// gitBisectUsageOutput is the usage line git prints for an invalid bisect
// subcommand.
func gitBisectUsageOutput() string {
	return "usage: git bisect [help|start|bad|good|new|old" +
		"|terms|skip|next|reset|visualize|replay|log|run]"
}

func TestRuleGitBisectUsage(t *testing.T) {
	resetTestState(t)
	rule := Get("git_bisect_usage")
	if rule == nil {
		t.Fatal("git_bisect_usage rule is not registered")
	}

	for _, script := range []string{"git bisect strt", "git bisect rset", "git bisect goood"} {
		t.Run("match/"+script, func(t *testing.T) {
			assertMatch(t, rule, script, gitBisectUsageOutput())
		})
	}

	for _, script := range []string{"git bisect", "git bisect start", "git bisect good"} {
		t.Run("not_match/"+script, func(t *testing.T) {
			assertNotMatch(t, rule, script, "")
		})
	}
	assertNotMatch(t, rule, "ls bisect", gitBisectUsageOutput())

	cases := []struct {
		script string
		want   []string
	}{
		{"git bisect goood", []string{"git bisect good", "git bisect old", "git bisect log"}},
		{"git bisect strt", []string{"git bisect start", "git bisect terms", "git bisect reset"}},
		{"git bisect rset", []string{"git bisect reset", "git bisect next", "git bisect start"}},
	}
	for _, testCase := range cases {
		t.Run("new_command/"+testCase.script, func(t *testing.T) {
			assertNewCommands(t, rule, testCase.script, gitBisectUsageOutput(), testCase.want)
		})
	}
}
