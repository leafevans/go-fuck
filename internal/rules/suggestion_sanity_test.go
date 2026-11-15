package rules

import (
	"strings"
	"testing"

	"github.com/leafevans/go-fuck/v2/internal/types"
)

// TestRulesDoNotSuggestTheOriginalCommand guards a class of useless rules: a
// rule that matches a command and then suggests the very same command would
// make --repeat run it again in a loop. Every suggestion has to change
// something.
func TestRulesDoNotSuggestTheOriginalCommand(t *testing.T) {
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
				if strings.TrimSpace(suggestion) == "" {
					t.Errorf("rule %s suggests an empty command for %q", rule.Name, corpusCase.script)
				}
				if strings.TrimSpace(suggestion) == strings.TrimSpace(corpusCase.script) {
					t.Errorf("rule %s suggests the original command %q for script %q output %q", rule.Name, suggestion, corpusCase.script, corpusCase.output)
				}
			}
		}
	}
}
