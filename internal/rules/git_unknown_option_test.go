package rules

import "testing"

// gitUnknownOptionPushOutput is the output git prints for a misspelled long
// option: the trace line, the error and the usage list with the valid options.
const gitUnknownOptionPushOutput = "14:59:10.717575 git.c:506               trace: built-in: git push --set-upstrem origin main\n" +
	"error: unknown option `set-upstrem'\n" +
	"usage: git push [<options>] [<repository> [<refspec>...]]\n" +
	"\n" +
	"    -v, --[no-]verbose    be more verbose\n" +
	"    -q, --[no-]quiet      be more quiet\n" +
	"    --[no-]repo <repository>\n" +
	"    --[no-]set-upstream   change the upstream info\n" +
	"    -u, --[no-]set-upstream\n" +
	"    --[no-]force          force updates\n"

func TestRuleGitUnknownOption(t *testing.T) {
	resetTestState(t)
	rule := Get("git_unknown_option")
	if rule == nil {
		t.Fatal("git_unknown_option rule is not registered")
	}

	assertMatch(t, rule, "git push --set-upstrem origin main", gitUnknownOptionPushOutput)
	assertNewCommands(t, rule, "git push --set-upstrem origin main", gitUnknownOptionPushOutput,
		[]string{"git push --set-upstream origin main"})

	// The rejected option has to be part of the command line.
	assertNotMatch(t, rule, "git status", gitUnknownOptionPushOutput)
	// Without the usage list there are no candidates.
	assertNotMatch(t, rule, "git push --set-upstrem origin main", "error: unknown option `set-upstrem'")
	// Short options are not handled by this rule.
	assertNotMatch(t, rule, "git push -x", "error: unknown switch `x'\nusage: git push [<options>]\n    -v, --verbose\n")

	assertNewCommands(t, rule, "git push --set-upstrem origin main", "error: unknown option `set-upstrem'",
		nil)
}
