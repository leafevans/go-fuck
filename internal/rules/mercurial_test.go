package rules

import (
	"slices"
	"testing"
)

func TestRuleMercurial(t *testing.T) {
	resetTestState(t)
	rule := Get("mercurial")

	t.Run("match", func(t *testing.T) {
		commands := []string{
			"hg: unknown command 'base'\n(did you mean one of blame, phase, rebase?)",
			"hg: unknown command 'branchch'\n(did you mean one of branch, branches?)",
			"hg: unknown command 'vert'\n(did you mean one of revert?)",
			"hg: command 're' is ambiguous:\n(did you mean one of log?)",
			"hg: unknown command 'rerere'\n(did you mean one of revert?)",
			"hg: command 're' is ambiguous:\n    rebase recover remove rename resolve revert",
		}
		scripts := []string{"hg base", "hg branchch", "hg vert", "hg lgo -r tip", "hg rerere", "hg re"}
		for index, script := range scripts {
			assertMatch(t, rule, script, commands[index])
		}
	})

	t.Run("not match", func(t *testing.T) {
		assertNotMatch(t, rule, "hg", "\nMercurial Distributed SCM\n\nbasic commands:")
		assertNotMatch(t, rule, "hg asdf", "hg: unknown command 'asdf'\nMercurial Distributed SCM\n\nbasic commands:")
		assertNotMatch(t, rule, "hg qwer", "hg: unknown command 'qwer'\nMercurial Distributed SCM\n\nbasic commands:")
		assertNotMatch(t, rule, "hg me", "\nabort: no repository found in './thefuck' (.hg not found)!")
		assertNotMatch(t, rule, "hg reb", "\nabort: no repository found in './thefuck' (.hg not found)!")
		assertNotMatch(t, rule, "hg co", "\nabort: no repository found in './thefuck' (.hg not found)!")
	})

	t.Run("extract possibilities", func(t *testing.T) {
		cases := []struct {
			command       string
			output        string
			possibilities []string
		}{
			{"hg base", "hg: unknown command 'base'\n(did you mean one of blame, phase, rebase?)",
				[]string{"blame", "phase", "rebase"}},
			{"hg branchch", "hg: unknown command 'branchch'\n(did you mean one of branch, branches?)",
				[]string{"branch", "branches"}},
			{"hg vert", "hg: unknown command 'vert'\n(did you mean one of revert?)",
				[]string{"revert"}},
			{"hg lgo -r tip", "hg: command 're' is ambiguous:\n(did you mean one of log?)",
				[]string{"log"}},
			{"hg rerere", "hg: unknown command 'rerere'\n(did you mean one of revert?)",
				[]string{"revert"}},
			{"hg re", "hg: command 're' is ambiguous:\n    rebase recover remove rename resolve revert",
				[]string{"rebase", "recover", "remove", "rename", "resolve", "revert"}},
			{"hg re re", "hg: command 're' is ambiguous:\n    rebase recover remove rename resolve revert",
				[]string{"rebase", "recover", "remove", "rename", "resolve", "revert"}},
		}
		for _, testCase := range cases {
			command := testCommand(testCase.command, testCase.output)
			if got := mercurialExtractPossibilities(command); !slices.Equal(got, testCase.possibilities) {
				t.Errorf("mercurialExtractPossibilities(%q) = %q, want %q",
					testCase.command, got, testCase.possibilities)
			}
		}
	})

	t.Run("get new command", func(t *testing.T) {
		assertNewCommands(t, rule, "hg base", "hg: unknown command 'base'\n(did you mean one of blame, phase, rebase?)",
			[]string{"hg rebase"})
		assertNewCommands(t, rule, "hg branchch", "hg: unknown command 'branchch'\n(did you mean one of branch, branches?)",
			[]string{"hg branch"})
		assertNewCommands(t, rule, "hg vert", "hg: unknown command 'vert'\n(did you mean one of revert?)",
			[]string{"hg revert"})
		assertNewCommands(t, rule, "hg lgo -r tip", "hg: command 're' is ambiguous:\n(did you mean one of log?)",
			[]string{"hg log -r tip"})
		assertNewCommands(t, rule, "hg rerere", "hg: unknown command 'rerere'\n(did you mean one of revert?)",
			[]string{"hg revert"})
		assertNewCommands(t, rule, "hg re", "hg: command 're' is ambiguous:\n    rebase recover remove rename resolve revert",
			[]string{"hg rebase"})
		assertNewCommands(t, rule, "hg re re", "hg: command 're' is ambiguous:\n    rebase recover remove rename resolve revert",
			[]string{"hg rebase re"})
	})
}
