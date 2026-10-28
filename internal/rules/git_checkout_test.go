package rules

import "testing"

// gitCheckoutDidNotMatch builds the output git prints for an unknown branch or
// file.
func gitCheckoutDidNotMatch(target string, didYouForget bool) string {
	output := "error: pathspec '" + target + "' did not match any file(s) known to git."
	if didYouForget {
		output += "\nDid you forget to 'git add'?'"
	}
	return output
}

func TestRuleGitCheckout(t *testing.T) {
	resetTestState(t)
	rule := Get("git_checkout")
	if rule == nil {
		t.Fatal("git_checkout rule is not registered")
	}

	original := gitCheckoutBranches
	t.Cleanup(func() { gitCheckoutBranches = original })

	assertMatch(t, rule, "git checkout unknown", gitCheckoutDidNotMatch("unknown", false))
	assertMatch(t, rule, "git commit unknown", gitCheckoutDidNotMatch("unknown", false))

	assertNotMatch(t, rule, "git submodule update unknown", gitCheckoutDidNotMatch("unknown", true))
	assertNotMatch(t, rule, "git checkout known", "")
	assertNotMatch(t, rule, "git commit known", "")

	branchCases := []struct {
		branches string
		want     []string
	}{
		{"", []string{}},
		{"* master", []string{"master"}},
		{"  remotes/origin/master", []string{"master"}},
		{"  remotes/origin/test/1", []string{"test/1"}},
		{"  remotes/origin/test/1/2/3", []string{"test/1/2/3"}},
		{"  test/1", []string{"test/1"}},
		{"  test/1/2/3", []string{"test/1/2/3"}},
		{"  remotes/origin/HEAD -> origin/master", []string{}},
		{"  just-another-branch", []string{"just-another-branch"}},
		{"* master\n  just-another-branch", []string{"master", "just-another-branch"}},
		{"* master\n  remotes/origin/master\n  just-another-branch",
			[]string{"master", "master", "just-another-branch"}},
	}
	for _, testCase := range branchCases {
		t.Run("get_branches/"+testCase.branches, func(t *testing.T) {
			got := gitCheckoutParseBranches(testCase.branches)
			if len(got) != len(testCase.want) {
				t.Fatalf("gitCheckoutParseBranches(%q) = %q, want %q",
					testCase.branches, got, testCase.want)
			}
			for index := range testCase.want {
				if got[index] != testCase.want[index] {
					t.Fatalf("gitCheckoutParseBranches(%q) = %q, want %q",
						testCase.branches, got, testCase.want)
				}
			}
		})
	}

	commandCases := []struct {
		branches string
		script   string
		target   string
		want     []string
	}{
		{"", "git checkout unknown", "unknown", []string{"git checkout -b unknown"}},
		{"", "git commit unknown", "unknown",
			[]string{"git branch unknown && git commit unknown"}},
		{"  test-random-branch-123", "git checkout tst-rdm-brnch-123", "tst-rdm-brnch-123",
			[]string{"git checkout test-random-branch-123", "git checkout -b tst-rdm-brnch-123"}},
		{"  test-random-branch-123", "git commit tst-rdm-brnch-123", "tst-rdm-brnch-123",
			[]string{"git commit test-random-branch-123"}},
	}
	for _, testCase := range commandCases {
		branches := testCase.branches
		gitCheckoutBranches = func() []string { return gitCheckoutParseBranches(branches) }
		t.Run("new_command/"+testCase.script, func(t *testing.T) {
			assertNewCommands(t, rule, testCase.script, gitCheckoutDidNotMatch(testCase.target, false),
				testCase.want)
		})
	}
}
