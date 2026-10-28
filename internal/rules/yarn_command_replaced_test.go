package rules

import "testing"

// yarnCommandReplacedOutput returns the output yarn prints for a replaced
// command.
func yarnCommandReplacedOutput(name string) string {
	return "error `install` has been replaced with `add` to add new " +
		"dependencies. Run \"yarn add " + name + "\" instead."
}

func TestRuleYarnCommandReplaced(t *testing.T) {
	resetTestState(t)

	rule := Get("yarn_command_replaced")
	if rule == nil {
		t.Fatal("yarn_command_replaced rule is not registered")
	}

	for _, name := range []string{"redux", "moment", "lodash"} {
		script := "yarn install " + name
		assertMatch(t, rule, script, yarnCommandReplacedOutput(name))
		assertNewCommands(t, rule, script, yarnCommandReplacedOutput(name),
			[]string{"yarn add " + name})
	}

	assertNotMatch(t, rule, "yarn install", "")
	assertNotMatch(t, rule, "npm install redux", yarnCommandReplacedOutput("redux"))
}
