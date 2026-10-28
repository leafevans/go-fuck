package rules

import "testing"

// gitFixStashOutput is the usage output git prints for an unknown stash
// subcommand.
func gitFixStashOutput() string {
	return "\nusage: git stash list [<options>]\n" +
		"   or: git stash show [<stash>]\n" +
		"   or: git stash drop [-q|--quiet] [<stash>]\n" +
		"   or: git stash ( pop | apply ) [--index] [-q|--quiet] [<stash>]\n" +
		"   or: git stash branch <branchname> [<stash>]\n" +
		"   or: git stash [save [--patch] [-k|--[no-]keep-index] [-q|--quiet]\n" +
		"\t\t       [-u|--include-untracked] [-a|--all] [<message>]]\n" +
		"   or: git stash clear\n"
}

func TestRuleGitFixStash(t *testing.T) {
	resetTestState(t)
	rule := Get("git_fix_stash")
	if rule == nil {
		t.Fatal("git_fix_stash rule is not registered")
	}

	for _, script := range []string{
		"git stash opp",
		"git stash Some message",
		"git stash saev Some message",
	} {
		t.Run("match/"+script, func(t *testing.T) {
			assertMatch(t, rule, script, gitFixStashOutput())
		})
	}
	assertNotMatch(t, rule, "git", gitFixStashOutput())

	commandCases := []struct {
		script string
		want   string
	}{
		{"git stash opp", "git stash pop"},
		{"git stash Some message", "git stash save Some message"},
		{"git stash saev Some message", "git stash save Some message"},
	}
	for _, testCase := range commandCases {
		t.Run("new_command/"+testCase.script, func(t *testing.T) {
			assertNewCommands(t, rule, testCase.script, gitFixStashOutput(), []string{testCase.want})
		})
	}
}
