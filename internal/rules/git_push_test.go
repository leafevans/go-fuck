package rules

import "testing"

// gitPushOutput returns the output git prints when a branch has no upstream.
func gitPushOutput(branch string) string {
	return "fatal: The current branch " + branch + " has no upstream branch.\n" +
		"To push the current branch and set the remote as upstream, use\n\n" +
		"    git push --set-upstream origin " + branch + "\n\n"
}

// gitPushBitbucketOutput returns the output of a successful push to Bitbucket.
func gitPushBitbucketOutput() string {
	return "To git@bitbucket.org:test.git\n" +
		"   e5e7fbb..700d998  feature/set-upstream -> feature/set-upstream\n" +
		"Branch feature/set-upstream set up to track remote branch " +
		"feature/set-upstream from origin.\n"
}

func TestRuleGitPush(t *testing.T) {
	rule := Get("git_push")
	if rule == nil {
		t.Fatal("git_push rule is not registered")
	}

	assertMatch(t, rule, "git push", gitPushOutput("master"))
	assertMatch(t, rule, "git push origin", gitPushOutput("master"))

	assertNotMatch(t, rule, "git push origin", gitPushBitbucketOutput())
	assertNotMatch(t, rule, "git push master", "")
	assertNotMatch(t, rule, "ls", gitPushOutput("master"))

	cases := []struct {
		script string
		branch string
		want   string
	}{
		{"git push", "master", "git push --set-upstream origin master"},
		{"git push master", "master", "git push --set-upstream origin master"},
		{"git push -u", "master", "git push --set-upstream origin master"},
		{"git push -u origin", "master", "git push --set-upstream origin master"},
		{"git push origin", "master", "git push --set-upstream origin master"},
		{"git push --set-upstream origin", "master", "git push --set-upstream origin master"},
		{"git push --quiet", "master", "git push --set-upstream origin master --quiet"},
		{"git push --quiet origin", "master", "git push --set-upstream origin master --quiet"},
		{"git -c test=test push --quiet origin", "master", "git -c test=test push --set-upstream origin master --quiet"},
		{"git push", "test's", "git push --set-upstream origin test" + backslash + "'s"},
		{"git push --force", "master", "git push --set-upstream origin master --force"},
		{"git push --force-with-lease", "master", "git push --set-upstream origin master --force-with-lease"},
	}
	for _, testCase := range cases {
		assertNewCommands(t, rule, testCase.script, gitPushOutput(testCase.branch),
			[]string{testCase.want})
	}
}
