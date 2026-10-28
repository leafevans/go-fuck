package rules

import "testing"

// gulpNotTaskOutput returns the output gulp prints for an unknown task.
func gulpNotTaskOutput(task string) string {
	return "[00:41:11] Using gulpfile gulpfile.js\n" +
		"[00:41:11] Task '" + task + "' is not in your gulpfile\n" +
		"[00:41:11] Please check the documentation for proper gulpfile formatting\n"
}

// gulpNotTaskStubTasks replaces the gulp task lookup with the fixture.
func gulpNotTaskStubTasks(t *testing.T) {
	t.Helper()
	original := gulpTasks
	gulpTasks = func() []string {
		return gulpNotTaskParseTasks("serve \nbuild \ndefault \n")
	}
	t.Cleanup(func() { gulpTasks = original })
}

func TestRuleGulpNotTask(t *testing.T) {
	resetTestState(t)
	gulpNotTaskStubTasks(t)

	rule := Get("gulp_not_task")
	if rule == nil {
		t.Fatal("gulp_not_task rule is not registered")
	}

	assertMatch(t, rule, "gulp srve", gulpNotTaskOutput("srve"))

	assertNotMatch(t, rule, "gulp serve", "")
	assertNotMatch(t, rule, "cat srve", gulpNotTaskOutput("srve"))

	assertNewCommands(t, rule, "gulp srve", gulpNotTaskOutput("srve"),
		[]string{"gulp serve", "gulp default"})
}
