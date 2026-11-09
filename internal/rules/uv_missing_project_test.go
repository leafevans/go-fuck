package rules

import "testing"

// uvMissingProjectOutput is the output of uv when no project file exists.
const uvMissingProjectOutput = "error: No \x60pyproject.toml\x60 found in current directory or any parent directory\n"

func TestRuleUvMissingProject(t *testing.T) {
	resetTestState(t)
	rule := Get("uv_missing_project")
	if rule == nil {
		t.Fatal("uv_missing_project rule is not registered")
	}

	assertMatch(t, rule, "uv add requests", uvMissingProjectOutput)
	assertMatch(t, rule, "uv sync", uvMissingProjectOutput)
	assertNotMatch(t, rule, "uv pip install requests", uvMissingProjectOutput)
	assertNotMatch(t, rule, "uv sync", "")

	assertNewCommands(t, rule, "uv add requests", uvMissingProjectOutput,
		[]string{"uv init && uv add requests"})
}
