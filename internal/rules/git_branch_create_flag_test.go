package rules

import (
	"testing"

	"github.com/leafevans/go-fuck/v2/internal/shells"
)

func TestRuleGitBranchCreateFlag(t *testing.T) {
	resetTestState(t)
	rule := Get("git_branch_create_flag")
	if rule == nil {
		t.Fatal("git_branch_create_flag rule is not registered")
	}
	const switchError = "error: unknown switch `b'\nusage: git switch [<options>] [<branch>]\n"
	const checkoutError = "error: unknown switch `c'\r\nusage: git checkout [<options>] <branch>\r\n"

	for _, tc := range []struct{ name, script, output, want string }{
		{"switch", "git switch -b feature/demo", switchError, "git switch -c feature/demo"},
		{"checkout", "git checkout -c feature/demo", checkoutError, "git checkout -b feature/demo"},
		{"preserve arguments", "git -C 'my repo' -c color.ui=false switch --quiet\t'-b'  'feature/demo' 'origin/main'", switchError,
			"git -C 'my repo' -c color.ui=false switch --quiet\t-c  'feature/demo' 'origin/main'"},
		{"unrelated executable", "echo checkout -c feature/demo", checkoutError, ""},
		{"unrelated subcommand", "git commit -c feature/demo", checkoutError, ""},
		{"unrelated diagnostic", "git switch -b feature/demo", "error: unknown switch `x'", ""},
		{"diagnostic not at line start", "git switch -b feature/demo", "message: " + switchError, ""},
		{"flag absent", "git switch -c feature/demo", switchError, ""},
		{"flag only global", "git -C -b switch feature/demo", switchError, ""},
		{"flag is an option value", "git switch --conflict -b -qb feature/demo", switchError, ""},
		{"bundled flag", "git switch -qb feature/demo", switchError, ""},
		{"missing branch", "git switch -b", switchError, ""},
		{"empty branch", "git switch -b ''", switchError, ""},
		{"flag instead of branch", "git switch -b -q", switchError, ""},
		{"flag after separator", "git checkout -- -c feature/demo", checkoutError, ""},
		{"path checkout", "git checkout -c feature/demo -- file", checkoutError, ""},
		{"replacement already present", "git checkout -c feature/demo -b other", checkoutError, ""},
		{"duplicate flag", "git switch -b feature/demo -b other", switchError, ""},
		{"compound command", "git switch -b feature/demo; echo done", switchError, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if tc.want == "" {
				assertNotMatch(t, rule, tc.script, tc.output)
				assertNewCommands(t, rule, tc.script, tc.output, nil)
				return
			}
			assertMatch(t, rule, tc.script, tc.output)
			assertNewCommands(t, rule, tc.script, tc.output, []string{tc.want})
		})
	}

	t.Run("PowerShell", func(t *testing.T) {
		shells.Set(shells.NewPowershell())
		t.Cleanup(func() { shells.Set(shells.NewGeneric()) })
		const script = `git -C 'C:\dev\my repo' switch -q -b 'feature/demo' 'origin/main'`
		assertMatch(t, rule, script, switchError)
		assertNewCommands(t, rule, script, switchError,
			[]string{`git -C 'C:\dev\my repo' switch -q -c 'feature/demo' 'origin/main'`})
	})
}
