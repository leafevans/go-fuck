package rules

import "testing"

func TestRuleFixAltSpace(t *testing.T) {
	resetTestState(t)
	rule := Get("fix_alt_space")
	if rule == nil {
		t.Fatal("fix_alt_space rule is not registered")
	}

	altSpace := string(rune(0x00A0))
	script := "ps -ef |" + altSpace + "grep foo"
	output := "-bash: " + altSpace + "grep: command not found"

	assertMatch(t, rule, script, output)
	assertNotMatch(t, rule, "ps -ef | grep foo", "")
	assertNotMatch(t, rule, "", "")

	assertNewCommands(t, rule, script, "", []string{"ps -ef | grep foo"})
}
