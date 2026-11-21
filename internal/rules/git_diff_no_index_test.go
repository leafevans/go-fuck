package rules

import "testing"

func TestRuleGitDiffNoIndex(t *testing.T) {
	resetTestState(t)
	rule := Get("git_diff_no_index")
	if rule == nil {
		t.Fatal("git_diff_no_index rule is not registered")
	}

	assertMatch(t, rule, "git diff foo bar", "")
	assertMatch(t, rule, "git -C 'repo dir' diff --name-only -- --no-index 'second file'", "")

	for _, script := range []string{
		"git diff --no-index foo bar",
		"git diff foo",
		"git diff foo bar baz",
		"git -C repo diff file",
		"git diff --output result.txt file",
		"git diff --cached one two",
	} {
		t.Run("not_match/"+script, func(t *testing.T) {
			assertNotMatch(t, rule, script, "")
		})
	}

	assertNewCommands(t, rule, "git diff foo bar", "",
		[]string{"git diff --no-index foo bar"})
}
