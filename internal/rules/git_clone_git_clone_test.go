package rules

import "testing"

// gitCloneGitCloneOutput is the output git prints for too many arguments.
func gitCloneGitCloneOutput() string {
	return "\nfatal: Too many arguments.\n\n" +
		"usage: git clone [<options>] [--] <repo> [<dir>]\n"
}

func TestRuleGitCloneGitClone(t *testing.T) {
	resetTestState(t)
	rule := Get("git_clone_git_clone")
	if rule == nil {
		t.Fatal("git_clone_git_clone rule is not registered")
	}

	assertMatch(t, rule, "git clone git clone foo", gitCloneGitCloneOutput())

	for _, script := range []string{"", "git branch", "git clone foo", "git clone foo bar baz"} {
		t.Run("not_match/"+script, func(t *testing.T) {
			assertNotMatch(t, rule, script, gitCloneGitCloneOutput())
		})
	}

	assertNewCommands(t, rule, "git clone git clone foo", gitCloneGitCloneOutput(),
		[]string{"git clone foo"})
}
