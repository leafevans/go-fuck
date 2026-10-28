package rules

import "testing"

// ghUnknownFlagOutput is the output of gh 2.95 for an unknown flag; the usage
// output lists the flags of the command.
func ghUnknownFlagOutput() string {
	return "unknown flag: --stat\n" +
		"\n" +
		"Usage:  gh pr list [flags]\n" +
		"Flags:\n" +
		"      --app string        Filter by GitHub App author\n" +
		"  -a, --assignee string   Filter by assignee\n" +
		"  -B, --base string       Filter by base branch\n" +
		"  -s, --state string      Filter by state: {open|closed|merged|all} (default \"open\")\n"
}

func TestRuleGhUnknownFlag(t *testing.T) {
	resetTestState(t)
	rule := Get("gh_unknown_flag")
	if rule == nil {
		t.Fatal("gh_unknown_flag rule is not registered")
	}

	assertMatch(t, rule, "gh pr list --stat", ghUnknownFlagOutput())
	assertNotMatch(t, rule, "git pr list --stat", ghUnknownFlagOutput())
	assertNotMatch(t, rule, "gh pr list --state open", "1\topen\tfix typo")

	assertNewCommands(t, rule, "gh pr list --stat", ghUnknownFlagOutput(),
		[]string{"gh pr list --state"})
}
