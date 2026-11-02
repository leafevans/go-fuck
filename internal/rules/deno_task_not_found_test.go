package rules

import "testing"

// denoTaskNotFoundOutput returns the output of deno for a missing task.
func denoTaskNotFoundOutput(task string) string {
	return "error: Task not found: " + task + "\n" +
		"Available tasks:\n" +
		"- build\n" +
		"    deno run -A build.ts\n" +
		"- lint\n" +
		"    // Lint the project\n" +
		"    deno lint\n" +
		"- test\n" +
		"    deno test\n"
}

func TestRuleDenoTaskNotFound(t *testing.T) {
	resetTestState(t)
	rule := Get("deno_task_not_found")
	if rule == nil {
		t.Fatal("deno_task_not_found rule is not registered")
	}

	assertMatch(t, rule, "deno task buidl", denoTaskNotFoundOutput("buidl"))
	assertNotMatch(t, rule, "just buidl", denoTaskNotFoundOutput("buidl"))
	assertNotMatch(t, rule, "deno task build", "error: something else\n")
	// Without the list of available tasks there is nothing to correct.
	assertNotMatch(t, rule, "deno task buidl", "error: Task not found: buidl\n")

	got, err := rule.GetNewCommand(testCommand("deno task buidl", denoTaskNotFoundOutput("buidl")))
	if err != nil {
		t.Fatalf("rule %q returned an error: %v", rule.Name, err)
	}
	if len(got) == 0 || got[0] != "deno task build" {
		t.Fatalf("rule %q returned %q, want first command %q", rule.Name, got, "deno task build")
	}
}

func TestRuleDenoTaskNotFoundWorkspaceHeading(t *testing.T) {
	resetTestState(t)
	rule := Get("deno_task_not_found")

	output := "\x1b[32mAvailable tasks (app):\x1b[0m\n" +
		"- \x1b[36mbuild\x1b[0m (workspace)\n" +
		"    deno run -A build.ts\n" +
		"- test\n" +
		"    deno test\n"
	got, err := rule.GetNewCommand(testCommand("deno task buidl", "error: Task not found: buidl\n"+output))
	if err != nil {
		t.Fatalf("rule %q returned an error: %v", rule.Name, err)
	}
	if len(got) == 0 || got[0] != "deno task build" {
		t.Fatalf("rule %q returned %q, want first command %q", rule.Name, got, "deno task build")
	}
}
