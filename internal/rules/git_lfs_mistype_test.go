package rules

import "testing"

// gitLfsMistypeOutput is the output git-lfs prints for a mistyped command.
func gitLfsMistypeOutput() string {
	return `
Error: unknown command "evn" for "git-lfs"

Did you mean this?
        env
        ext

Run 'git-lfs --help' for usage.
    `
}

func TestRuleGitLfsMistype(t *testing.T) {
	resetTestState(t)
	rule := Get("git_lfs_mistype")
	if rule == nil {
		t.Fatal("git_lfs_mistype rule is not registered")
	}

	assertMatch(t, rule, "git lfs evn", gitLfsMistypeOutput())
	assertNotMatch(t, rule, "git lfs env", "bash: git: command not found")
	assertNotMatch(t, rule, "docker lfs env", gitLfsMistypeOutput())

	assertNewCommands(t, rule, "git lfs evn", gitLfsMistypeOutput(),
		[]string{"git lfs env", "git lfs ext"})
}
