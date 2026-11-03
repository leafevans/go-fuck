package rules

import "testing"

// taskUnknownTaskOutput returns the output of go-task for a misspelled task.
func taskUnknownTaskOutput(task, suggestion string) string {
	return "task: Task \"" + task + "\" does not exist. Did you mean \"" + suggestion + "\"?\n"
}

func TestRuleTaskUnknownTask(t *testing.T) {
	resetTestState(t)
	rule := Get("task_unknown_task")
	if rule == nil {
		t.Fatal("task_unknown_task rule is not registered")
	}

	assertMatch(t, rule, "task instal", taskUnknownTaskOutput("instal", "install"))
	assertMatch(t, rule, "task --verbose instal", taskUnknownTaskOutput("instal", "install"))
	assertNotMatch(t, rule, "make instal", taskUnknownTaskOutput("instal", "install"))
	// Without a suggestion there is nothing to correct.
	assertNotMatch(t, rule, "task instal", "task: Task \"instal\" does not exist\n")

	assertNewCommands(t, rule, "task instal", taskUnknownTaskOutput("instal", "install"), []string{"task install"})
}
