package rules

import "testing"

// uvFailedSpawnOutput is the output of uv run for a file that does not exist.
const uvFailedSpawnOutput = "error: Failed to spawn: mian.py\n  cause: program not found\n"

// stubUvScriptCandidates replaces the file list of the rule.
func stubUvScriptCandidates(t *testing.T, names []string) {
	t.Helper()
	original := uvScriptCandidates
	uvScriptCandidates = func() []string { return names }
	t.Cleanup(func() { uvScriptCandidates = original })
}

func TestRuleUvScriptNotFound(t *testing.T) {
	resetTestState(t)
	rule := Get("uv_script_not_found")
	if rule == nil {
		t.Fatal("uv_script_not_found rule is not registered")
	}

	stubUvScriptCandidates(t, []string{"main.py", "lib.py"})
	assertMatch(t, rule, "uv run mian.py", uvFailedSpawnOutput)
	assertNotMatch(t, rule, "uv run mian.py", "")
	assertNotMatch(t, rule, "uv run mian.py", "error: Failed to spawn: mian.py\n")
	assertNotMatch(t, rule, "python mian.py", uvFailedSpawnOutput)
	assertNewCommands(t, rule, "uv run mian.py", uvFailedSpawnOutput,
		[]string{"uv run main.py"})
	assertNewCommands(t, rule, "uv run --no-project mian.py", uvFailedSpawnOutput,
		[]string{"uv run --no-project main.py"})

	stubUvScriptCandidates(t, []string{"main.py"})
	assertNotMatch(t, rule, "uv run pytest",
		"error: Failed to spawn: pytest\n  cause: program not found\n")
}
