package rules

import (
	"testing"

	"github.com/leafevans/go-fuck/internal/types"
)

// FuzzRules checks that no rule panics on an arbitrary command line or output.
//
// The rule execution in types.Rule recovers panics and reports a failed rule,
// so a panic does not crash go-fuck. It still means that the rule is broken, so
// the fuzz target calls the functions directly and fails on a panic.
func FuzzRules(f *testing.F) {
	seeds := []struct {
		script string
		output string
	}{
		{"git push", "fatal: The current branch main has no upstream branch."},
		{"rm dir", "Windows PowerShell is in NonInteractive mode."},
		{"cp a b", "cp: cannot stat a: No such file or directory"},
		{"mkdir dir", "File exists"},
		{"", ""},
		{"\x00\x01", "\xff\xfe"},
	}
	for _, seed := range seeds {
		f.Add(seed.script, seed.output)
	}
	for _, seed := range ruleCorpus() {
		f.Add(seed.script, seed.output)
	}
	registered := All()
	if len(registered) == 0 {
		f.Fatal("no rules are registered")
	}
	f.Fuzz(func(t *testing.T, script, output string) {
		command := types.NewCommand(script, output)
		for _, rule := range registered {
			rule := rule
			func() {
				defer func() {
					if recovered := recover(); recovered != nil {
						t.Fatalf("rule %q panicked on script %q and output %q: %v",
							rule.Name, script, output, recovered)
					}
				}()
				if rule.Match(command) {
					_, _ = rule.GetNewCommand(command)
				}
			}()
		}
	})
}
