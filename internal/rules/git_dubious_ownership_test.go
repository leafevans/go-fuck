package rules

import "testing"

// gitDubiousOwnershipOutput is the output that git prints for a repository
// owned by another user (see setup.c in git). Git shell quotes the path in the
// exemption command, so repository and quotedPath may differ.
func gitDubiousOwnershipOutput(repository, quotedPath string) string {
	return "fatal: detected dubious ownership in repository at '" + repository + "'\n" +
		"'" + repository + "' is owned by:\n" +
		"\t'S-1-5-21-1111111111-2222222222-3333333333-1001'\n" +
		"but the current user is:\n" +
		"\t'S-1-5-21-1111111111-2222222222-3333333333-1002'\n" +
		"To add an exception for this directory, call:\n" +
		"\n" +
		"\tgit config --global --add safe.directory " + quotedPath + "\n"
}

func TestRuleGitDubiousOwnership(t *testing.T) {
	resetTestState(t)
	rule := Get("git_dubious_ownership")
	if rule == nil {
		t.Fatal("git_dubious_ownership rule is not registered")
	}

	assertMatch(t, rule, "git status", gitDubiousOwnershipOutput("C:/repo", "C:/repo"))
	assertNotMatch(t, rule, "git status", "On branch main\nnothing to commit")
	assertNotMatch(t, rule, "svn status", gitDubiousOwnershipOutput("C:/repo", "C:/repo"))

	assertNewCommands(t, rule, "git status", gitDubiousOwnershipOutput("C:/repo", "C:/repo"),
		[]string{"git config --global --add safe.directory C:/repo && git status"})
	assertNewCommands(t, rule, "git log", gitDubiousOwnershipOutput("C:/my repo", "'C:/my repo'"),
		[]string{"git config --global --add safe.directory 'C:/my repo' && git log"})
}
