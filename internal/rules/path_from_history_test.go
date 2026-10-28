package rules

import (
	"testing"

	"github.com/leafevans/go-fuck/internal/types"
)

// pathFromHistoryFakeHistory returns a fixed history for the tests.
func pathFromHistoryFakeHistory(*types.Command) []string {
	return []string{"cd /opt/java", "ls ~/work/project/"}
}

func TestRulePathFromHistory(t *testing.T) {
	resetTestState(t)
	rule := Get("path_from_history")
	if rule == nil {
		t.Fatal("path_from_history rule is not registered")
	}

	originalHistory := pathFromHistoryGetValidHistory
	originalExists := pathFromHistoryExists
	pathFromHistoryGetValidHistory = pathFromHistoryFakeHistory
	pathFromHistoryExists = func(string) bool { return true }
	t.Cleanup(func() {
		pathFromHistoryGetValidHistory = originalHistory
		pathFromHistoryExists = originalExists
	})

	for _, testCase := range []struct {
		script string
		output string
	}{
		{"ls project", "no such file or directory: project"},
		{"cd project", "can't cd to project"},
	} {
		assertMatch(t, rule, testCase.script, testCase.output)
	}
	for _, testCase := range []struct {
		script string
		output string
	}{
		{"myapp cats", "no such file or directory: project"},
		{"cd project", ""},
	} {
		assertNotMatch(t, rule, testCase.script, testCase.output)
	}

	for _, testCase := range []struct {
		script string
		output string
		want   string
	}{
		{"ls project", "no such file or directory: project", "ls ~/work/project"},
		{"cd java", "can't cd to java", "cd /opt/java"},
	} {
		assertNewCommands(t, rule, testCase.script, testCase.output, []string{testCase.want})
	}
}
