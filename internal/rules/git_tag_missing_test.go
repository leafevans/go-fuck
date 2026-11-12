package rules

import "testing"

// stubGitTagNames replaces the tag list of the rule.
func stubGitTagNames(t *testing.T, names []string) {
	t.Helper()
	original := gitTagNamesCommand
	gitTagNamesCommand = func() []string { return names }
	t.Cleanup(func() { gitTagNamesCommand = original })
}

func TestRuleGitTagMissing(t *testing.T) {
	resetTestState(t)
	rule := Get("git_tag_missing")
	if rule == nil {
		t.Fatal("git_tag_missing rule is not registered")
	}

	stubGitTagNames(t, []string{"v1.0.0", "v2.0.0"})
	assertMatch(t, rule, "git tag -d v1.0.0x", "error: tag 'v1.0.0x' not found.\n")
	assertMatch(t, rule, "git tag --delete v1.0.0x", "error: tag 'v1.0.0x' not found.\n")
	assertMatch(t, rule, "git tag -v v1.0.0x", "error: tag 'v1.0.0x' not found.\n")
	assertMatch(t, rule, "git tag --verify v1.0.0x", "error: tag 'v1.0.0x' not found.\n")
	assertMatch(t, rule, "git -C repo tag -d v1.0.0x", "error: tag 'v1.0.0x' not found.\n")

	assertNotMatch(t, rule, "git tag -d zzz", "error: tag 'zzz' not found.\n")
	assertNotMatch(t, rule, "git tag -a v1.0.0x", "error: tag 'v1.0.0x' not found.\n")
	assertNotMatch(t, rule, "git push origin v1.0.0x", "error: tag 'v1.0.0x' not found.\n")
	assertNotMatch(t, rule, "git tag -d v1.0.0x", "")

	assertNewCommands(t, rule, "git tag -d v1.0.0x", "error: tag 'v1.0.0x' not found.\n",
		[]string{"git tag -d v1.0.0"})
	assertNewCommands(t, rule, "git tag -v v1.0.0x", "error: tag 'v1.0.0x' not found.\n",
		[]string{"git tag -v v1.0.0"})
}
