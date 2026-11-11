package rules

import "testing"

func TestRuleCmdPushdMkdir(t *testing.T) {
	resetTestState(t)
	rule := Get("cmd_pushd_mkdir")
	if rule == nil {
		t.Fatal("cmd_pushd_mkdir rule is not registered")
	}

	assertMatch(t, rule, "pushd C:\\missing", cmdPushdPathNotFound)
	assertMatch(t, rule, "pushd \"C:\\Program Files\\missing\"", cmdPushdPathNotFound)
	assertNotMatch(t, rule, "pushd C:\\missing", "")
	assertNotMatch(t, rule, "cd C:\\missing", cmdPushdPathNotFound)
	assertNotMatch(t, rule, "pushd", cmdPushdPathNotFound)

	assertNewCommands(t, rule, "pushd C:\\missing", cmdPushdPathNotFound,
		[]string{"mkdir \"C:\\missing\" && pushd \"C:\\missing\""})
}
