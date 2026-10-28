package rules

import "testing"

// gitAddForceOutput is the output git prints for ignored files.
func gitAddForceOutput() string {
	return "The following paths are ignored by one of your .gitignore files:\n" +
		"dist/app.js\n" +
		"dist/background.js\n" +
		"dist/options.js\n" +
		"Use -f if you really want to add them.\n"
}

func TestRuleGitAddForce(t *testing.T) {
	resetTestState(t)
	rule := Get("git_add_force")
	if rule == nil {
		t.Fatal("git_add_force rule is not registered")
	}

	assertMatch(t, rule, "git add dist/*.js", gitAddForceOutput())
	assertNotMatch(t, rule, "git add dist/*.js", "")

	assertNewCommands(t, rule, "git add dist/*.js", gitAddForceOutput(),
		[]string{"git add --force dist/*.js"})
}
