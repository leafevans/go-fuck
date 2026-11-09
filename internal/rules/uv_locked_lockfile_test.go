package rules

import "testing"

// uvLockedLockfileOutput is the uv output for a stale lock file and --locked.
func uvLockedLockfileOutput() string {
	return "error: The lockfile at `uv.lock` needs to be updated, but `--locked` was provided.\n" +
		"hint: To update the lockfile, run `uv lock`.\n"
}

func TestRuleUvLockedLockfile(t *testing.T) {
	resetTestState(t)
	rule := Get("uv_locked_lockfile")
	if rule == nil {
		t.Fatal("uv_locked_lockfile rule is not registered")
	}

	assertMatch(t, rule, "uv sync --locked", uvLockedLockfileOutput())
	assertMatch(t, rule, "uv run --locked script.py", uvLockedLockfileOutput())
	assertNotMatch(t, rule, "uv sync", uvLockedLockfileOutput())
	assertNotMatch(t, rule, "uv sync --locked", "Resolved 8 packages in 1.35s")
	assertNotMatch(t, rule, "pip sync --locked", uvLockedLockfileOutput())

	assertNewCommands(t, rule, "uv sync --locked", uvLockedLockfileOutput(),
		[]string{"uv sync"})
	assertNewCommands(t, rule, "uv run --locked script.py", uvLockedLockfileOutput(),
		[]string{"uv run script.py"})
}
