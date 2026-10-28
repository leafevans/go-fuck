package rules

import "testing"

func TestRuleCdMkdir(t *testing.T) {
	rule := Get("cd_mkdir")
	if rule == nil {
		t.Fatal("cd_mkdir rule is not registered")
	}

	for _, script := range []string{
		"cd foo",
		"cd foo/bar",
	} {
		assertMatch(t, rule, script, "cd: foo: No such file or directory")
	}
	assertMatch(t, rule, "cd foo", "cd: can't cd to foo")
	assertMatch(t, rule, "cd /foo/bar", "cd: The directory \"/foo/bar\" does not exist")

	assertNotMatch(t, rule, "cd foo", "")
	assertNotMatch(t, rule, "ls foo", "cd: foo: No such file or directory")

	assertNewCommands(t, rule, "cd foo", "cd: foo: No such file or directory",
		[]string{"mkdir -p foo && cd foo"})
	assertNewCommands(t, rule, "sudo cd foo", "cd: foo: No such file or directory",
		[]string{"sudo mkdir -p foo && cd foo"})
}
