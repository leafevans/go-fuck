package rules

import "testing"

// cmdRmdirNotEmptyOutput is the cmd.exe error for rmdir on a non-empty
// directory.
const cmdRmdirNotEmptyOutput = "The directory is not empty.\n"

func TestRuleCmdRmdirNonEmpty(t *testing.T) {
	resetTestState(t)
	rule := Get("cmd_rmdir_nonempty")
	if rule == nil {
		t.Fatal("cmd_rmdir_nonempty rule is not registered")
	}

	assertMatch(t, rule, "rmdir nonempty", cmdRmdirNotEmptyOutput)
	assertMatch(t, rule, "rd nonempty", cmdRmdirNotEmptyOutput)
	assertNotMatch(t, rule, "rmdir nonempty", "")
	assertNotMatch(t, rule, "del nonempty", cmdRmdirNotEmptyOutput)
	assertNotMatch(t, rule, "rmdir nonempty", "The system cannot find the path specified.\n")

	assertNewCommands(t, rule, "rmdir nonempty", cmdRmdirNotEmptyOutput,
		[]string{"rmdir /s /q nonempty"})
	assertNewCommands(t, rule, "rd nonempty", cmdRmdirNotEmptyOutput,
		[]string{"rd /s /q nonempty"})
}
