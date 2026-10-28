package rules

import "testing"

// leinNotTaskOutput is the lein output for an unknown task.
const leinNotTaskOutput = `'rpl' is not a task. See 'lein help'.

Did you mean this?
         repl
         jar
`

func TestRuleLeinNotTask(t *testing.T) {
	resetTestState(t)

	rule := Get("lein_not_task")
	if rule == nil {
		t.Fatal("lein_not_task rule is not registered")
	}

	assertMatch(t, rule, "lein rpl", leinNotTaskOutput)
	assertNotMatch(t, rule, "ls", leinNotTaskOutput)
	assertNotMatch(t, rule, "lein rpl", "")

	assertNewCommands(t, rule, "lein rpl --help", leinNotTaskOutput,
		[]string{"lein repl --help", "lein jar --help"})

	// The rule supports sudo like the original implementation.
	assertMatch(t, rule, "sudo lein rpl", leinNotTaskOutput)
	assertNewCommands(t, rule, "sudo lein rpl --help", leinNotTaskOutput,
		[]string{"sudo lein repl --help", "sudo lein jar --help"})
}
