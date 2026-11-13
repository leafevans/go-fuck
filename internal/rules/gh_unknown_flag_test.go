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

// ghUnknownFlagHelpOutput is the output of gh for a misspelled --help flag. The
// usage output does not list the help flag, although every command accepts it.
func ghUnknownFlagHelpOutput() string {
	return "unknown flag: --hlep\n" +
		"\n" +
		"Usage:  gh pr view [<number> | <url> | <branch>] [flags]\n" +
		"Flags:\n" +
		"  -c, --comments          View pull request comments\n" +
		"  -q, --jq expression     Filter JSON output using a jq expression\n" +
		"      --json fields       Output JSON with the specified fields\n" +
		"  -w, --web               Open a pull request in the browser\n"
}

// TestRuleGhUnknownFlagHelp checks that a typo of the help flag is corrected to
// --help even though the usage output does not list the flag.
func TestRuleGhUnknownFlagHelp(t *testing.T) {
	resetTestState(t)
	rule := Get("gh_unknown_flag")
	if rule == nil {
		t.Fatal("gh_unknown_flag rule is not registered")
	}

	assertMatch(t, rule, "gh pr view --hlep", ghUnknownFlagHelpOutput())
	assertNewCommands(t, rule, "gh pr view --hlep", ghUnknownFlagHelpOutput(),
		[]string{"gh pr view --help"})
}
