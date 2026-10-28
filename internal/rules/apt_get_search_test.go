package rules

import "testing"

func TestRuleAptGetSearch(t *testing.T) {
	resetTestState(t)
	rule := Get("apt_get_search")
	if rule == nil {
		t.Fatal("apt_get_search rule is not registered")
	}

	assertMatch(t, rule, "apt-get search foo", "")

	for _, script := range []string{
		"apt-cache search foo",
		"aptitude search foo",
		"apt search foo",
		"apt-get install foo",
		"apt-get source foo",
		"apt-get clean",
		"apt-get remove",
		"apt-get update",
	} {
		assertNotMatch(t, rule, script, "")
	}

	assertNewCommands(t, rule, "apt-get search foo", "",
		[]string{"apt-cache search foo"})
}
