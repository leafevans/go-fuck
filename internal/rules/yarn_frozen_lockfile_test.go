package rules

import "testing"

// yarnFrozenLockfileOutput is the yarn 1 output for a stale lock file and
// --frozen-lockfile.
func yarnFrozenLockfileOutput() string {
	return "yarn install v1.22.22\n" +
		"[1/4] Resolving packages...\n" +
		"error Your lockfile needs to be updated, but yarn was run with `--frozen-lockfile`.\n" +
		"info Visit https://yarnpkg.com/en/docs/cli/install for documentation about this command.\n"
}

func TestRuleYarnFrozenLockfile(t *testing.T) {
	resetTestState(t)
	rule := Get("yarn_frozen_lockfile")
	if rule == nil {
		t.Fatal("yarn_frozen_lockfile rule is not registered")
	}

	assertMatch(t, rule, "yarn install --frozen-lockfile", yarnFrozenLockfileOutput())
	assertMatch(t, rule, "corepack yarn install --frozen-lockfile", yarnFrozenLockfileOutput())
	assertNotMatch(t, rule, "npm install --frozen-lockfile", yarnFrozenLockfileOutput())
	assertNotMatch(t, rule, "yarn install", "")

	assertNewCommands(t, rule, "yarn install --frozen-lockfile", yarnFrozenLockfileOutput(),
		[]string{"yarn install --no-frozen-lockfile"})
	assertNewCommands(t, rule, "yarn install", yarnFrozenLockfileOutput(),
		[]string{"yarn install --no-frozen-lockfile"})
}
