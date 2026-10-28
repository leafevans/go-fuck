package rules

import (
	"testing"

	"github.com/leafevans/go-fuck/internal/types"
)

// historyFakeHistory returns a fixed history for the history rule tests.
func historyFakeHistory(*types.Command) []string {
	return []string{"ls cat", "diff x"}
}

func TestRuleHistory(t *testing.T) {
	resetTestState(t)
	rule := Get("history")
	if rule == nil {
		t.Fatal("history rule is not registered")
	}

	original := historyGetValidHistory
	historyGetValidHistory = historyFakeHistory
	t.Cleanup(func() { historyGetValidHistory = original })

	for _, script := range []string{"ls cet", "daff x"} {
		assertMatch(t, rule, script, "")
	}
	for _, script := range []string{"apt-get", "nocommand y"} {
		assertNotMatch(t, rule, script, "")
	}

	cases := []struct {
		script string
		want   string
	}{
		{"ls cet", "ls cat"},
		{"daff x", "diff x"},
	}
	for _, testCase := range cases {
		assertNewCommands(t, rule, testCase.script, "", []string{testCase.want})
	}
}
