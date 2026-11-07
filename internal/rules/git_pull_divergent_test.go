package rules

import "testing"

// gitPullDivergentOutput is the output of a pull that has to reconcile
// divergent branches without a configured strategy (git 2.27 and newer).
func gitPullDivergentOutput() string {
	return "From github.com:leafevans/go-fuck\n" +
		"   b7d3f0c..d41d8cd  main       -> origin/main\n" +
		"hint: You have divergent branches and need to specify how to reconcile them.\n" +
		"hint: You can do so by running one of the following commands sometime before\n" +
		"hint: your next pull:\n" +
		"hint:\n" +
		"hint:   git config pull.rebase false  # merge\n" +
		"hint:   git config pull.rebase true   # rebase\n" +
		"hint:   git config pull.ff only       # fast-forward only\n" +
		"hint:\n" +
		"hint: You can replace \"git config\" with \"git config --global\" to set a default\n" +
		"hint: preference for all repositories. You can also pass --rebase, --no-rebase,\n" +
		"hint: or --ff-only on the command line to override the configured default per\n" +
		"hint: invocation.\n" +
		"fatal: Need to specify how to reconcile divergent branches.\n"
}

func TestRuleGitPullDivergent(t *testing.T) {
	resetTestState(t)
	rule := Get("git_pull_divergent")
	if rule == nil {
		t.Fatal("git_pull_divergent rule is not registered")
	}

	assertMatch(t, rule, "git pull", gitPullDivergentOutput())
	assertMatch(t, rule, "git pull origin main", gitPullDivergentOutput())
	assertNotMatch(t, rule, "git push", gitPullDivergentOutput())
	assertNotMatch(t, rule, "git pull", "Already up to date.")

	assertNewCommands(t, rule, "git pull", gitPullDivergentOutput(),
		[]string{"git pull --no-rebase", "git pull --rebase"})
	assertNewCommands(t, rule, "git pull origin main", gitPullDivergentOutput(),
		[]string{"git pull --no-rebase origin main", "git pull --rebase origin main"})
}
