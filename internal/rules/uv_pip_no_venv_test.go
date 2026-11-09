package rules

import "testing"

// uvNoVirtualEnvOutput is the output of uv pip when no virtual environment
// exists.
const uvNoVirtualEnvOutput = "error: No virtual environment found; run \x60uv venv\x60 to create an environment, " +
	"or pass \x60--system\x60 to install into a non-virtual environment\n"

func TestRuleUvPipNoVenv(t *testing.T) {
	resetTestState(t)
	rule := Get("uv_pip_no_venv")
	if rule == nil {
		t.Fatal("uv_pip_no_venv rule is not registered")
	}

	assertMatch(t, rule, "uv pip install requests", uvNoVirtualEnvOutput)
	assertNotMatch(t, rule, "uv sync", uvNoVirtualEnvOutput)
	assertNotMatch(t, rule, "pip install requests", uvNoVirtualEnvOutput)
	assertNotMatch(t, rule, "uv pip install requests", "")

	assertNewCommands(t, rule, "uv pip install requests", uvNoVirtualEnvOutput,
		[]string{"uv venv && uv pip install requests"})
}
