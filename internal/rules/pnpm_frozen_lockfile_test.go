package rules

import "testing"

// pnpmFrozenLockfileOutput is the pnpm output for a stale lock file and
// --frozen-lockfile.
func pnpmFrozenLockfileOutput() string {
	return "Error: ERR_PNPM_OUTDATED_LOCKFILE\n" +
		"  × installing dependencies\n" +
		"  ╰─▶ Cannot install with \"frozen-lockfile\" because pnpm-lock.yaml is not up\n" +
		"      to date with package.json.\n"
}

func TestRulePnpmFrozenLockfile(t *testing.T) {
	resetTestState(t)
	rule := Get("pnpm_frozen_lockfile")
	if rule == nil {
		t.Fatal("pnpm_frozen_lockfile rule is not registered")
	}

	assertMatch(t, rule, "pnpm install --frozen-lockfile", pnpmFrozenLockfileOutput())
	assertMatch(t, rule, "corepack pnpm install --frozen-lockfile", pnpmFrozenLockfileOutput())
	assertNotMatch(t, rule, "npm install --frozen-lockfile", pnpmFrozenLockfileOutput())
	assertNotMatch(t, rule, "pnpm install", "")

	assertNewCommands(t, rule, "pnpm install --frozen-lockfile", pnpmFrozenLockfileOutput(),
		[]string{"pnpm install --no-frozen-lockfile"})
	assertNewCommands(t, rule, "pnpm install", pnpmFrozenLockfileOutput(),
		[]string{"pnpm install --no-frozen-lockfile"})
}
