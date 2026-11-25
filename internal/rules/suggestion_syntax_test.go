package rules

import (
	"os/exec"
	"runtime"
	"testing"

	"github.com/leafevans/go-fuck/v4/internal/types"
)

// TestRuleSuggestionsAreShellSyntax checks that every suggestion can be parsed
// by a POSIX shell. Rules build command lines by hand and a broken quote would
// make the correction fail after the user accepts it.
func TestRuleSuggestionsAreShellSyntax(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the syntax check needs a POSIX shell")
	}
	bash, err := exec.LookPath("bash")
	if err != nil {
		t.Skip("bash is not available")
	}
	resetTestState(t)
	for _, rule := range All() {
		for _, corpusCase := range ruleCorpus() {
			command := types.NewCommand(corpusCase.script, corpusCase.output)
			if !rule.Match(command) {
				continue
			}
			suggestions, err := rule.GetNewCommand(command)
			if err != nil {
				t.Fatalf("rule %s returned an error: %v", rule.Name, err)
			}
			for _, suggestion := range suggestions {
				if output, err := exec.Command(bash, "-n", "-c", suggestion).CombinedOutput(); err != nil {
					t.Errorf("rule %s suggests shell syntax error %q: %v\n%s", rule.Name, suggestion, err, output)
				}
			}
		}
	}
}
