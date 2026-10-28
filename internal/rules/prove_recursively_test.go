package rules

import "testing"

// proveRecursivelyOutput is the output of a prove run without tests.
const proveRecursivelyOutput = `Files=0, Tests=0,  0 wallclock secs ( 0.00 usr +  0.00 sys =  0.00 CPU)
Result: NOTESTS`

// proveRecursivelyUseIsDir replaces the directory check with a canned result.
func proveRecursivelyUseIsDir(t *testing.T, result bool) {
	t.Helper()
	original := proveRecursivelyIsDir
	proveRecursivelyIsDir = func(string) bool { return result }
	t.Cleanup(func() { proveRecursivelyIsDir = original })
}

func TestRuleProveRecursively(t *testing.T) {
	resetTestState(t)
	rule := Get("prove_recursively")

	t.Run("match", func(t *testing.T) {
		proveRecursivelyUseIsDir(t, true)
		assertMatch(t, rule, "prove -lv t", proveRecursivelyOutput)
		assertMatch(t, rule, "prove app/t", proveRecursivelyOutput)
	})

	t.Run("not match", func(t *testing.T) {
		proveRecursivelyUseIsDir(t, false)
		assertNotMatch(t, rule, "prove -lv t", proveRecursivelyOutput)

		proveRecursivelyUseIsDir(t, true)
		assertNotMatch(t, rule, "prove -r t", proveRecursivelyOutput)
		assertNotMatch(t, rule, "prove --recurse t", proveRecursivelyOutput)
	})

	t.Run("get new command", func(t *testing.T) {
		assertNewCommands(t, rule, "prove -lv t", proveRecursivelyOutput, []string{"prove -r -lv t"})
		assertNewCommands(t, rule, "prove t", proveRecursivelyOutput, []string{"prove -r t"})
	})
}
