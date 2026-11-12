package rules

import "testing"

// gitRemoteAlreadyExistsOutput is the git error for a duplicate remote.
func gitRemoteAlreadyExistsOutput(name string) string {
	return "error: remote " + name + " already exists.\n"
}

func TestRuleGitRemoteAlreadyExists(t *testing.T) {
	resetTestState(t)
	rule := Get("git_remote_already_exists")
	if rule == nil {
		t.Fatal("git_remote_already_exists rule is not registered")
	}

	assertMatch(t, rule, "git remote add origin https://example.com/second",
		gitRemoteAlreadyExistsOutput("origin"))
	assertMatch(t, rule, "git remote add upstream https://example.com/up2",
		gitRemoteAlreadyExistsOutput("upstream"))
	assertMatch(t, rule, "git -C repo remote add origin https://example.com/second",
		gitRemoteAlreadyExistsOutput("origin"))
	// Older git versions print "fatal:" instead of "error:".
	assertMatch(t, rule, "git remote add origin url",
		"fatal: remote origin already exists.\n")

	assertNotMatch(t, rule, "git remote set-url origin https://example.com/second",
		gitRemoteAlreadyExistsOutput("origin"))
	assertNotMatch(t, rule, "git remote add -f origin https://example.com/second",
		gitRemoteAlreadyExistsOutput("origin"))
	assertNotMatch(t, rule, "git remote add origin url", "fatal: not a git repository\n")
	assertNotMatch(t, rule, "git push origin main", gitRemoteAlreadyExistsOutput("origin"))

	assertNewCommands(t, rule, "git remote add origin https://example.com/second",
		gitRemoteAlreadyExistsOutput("origin"),
		[]string{"git remote set-url origin https://example.com/second"})
	assertNewCommands(t, rule, "git -C repo remote add origin https://example.com/second",
		gitRemoteAlreadyExistsOutput("origin"),
		[]string{"git -C repo remote set-url origin https://example.com/second"})
}
