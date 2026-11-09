package rules

import "testing"

// gitFetchClobberTagOutput is the output of git fetch --tags when a tag on the
// remote was rewritten.
func gitFetchClobberTagOutput() string {
	return "From github.com:leafevans/go-fuck\n" +
		"   cef33f7..3e7d93e  master     -> origin/master\n" +
		" ! [rejected]        v1         -> v1  (would clobber existing tag)\n"
}

func TestRuleGitFetchClobberTag(t *testing.T) {
	resetTestState(t)
	rule := Get("git_fetch_clobber_tag")
	if rule == nil {
		t.Fatal("git_fetch_clobber_tag rule is not registered")
	}

	assertMatch(t, rule, "git fetch --tags", gitFetchClobberTagOutput())
	assertMatch(t, rule, "git fetch origin", gitFetchClobberTagOutput())
	assertNotMatch(t, rule, "git push --tags", gitFetchClobberTagOutput())
	assertNotMatch(t, rule, "git fetch", "")

	assertNewCommands(t, rule, "git fetch --tags", gitFetchClobberTagOutput(),
		[]string{"git fetch --force --tags"})
	assertNewCommands(t, rule, "git fetch origin", gitFetchClobberTagOutput(),
		[]string{"git fetch --tags --force origin"})
	assertNewCommands(t, rule, "git fetch --tags --force", gitFetchClobberTagOutput(), nil)
}
