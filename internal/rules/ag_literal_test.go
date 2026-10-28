package rules

import "testing"

// agLiteralOutput returns the error that ag prints for a bad regular
// expression.
func agLiteralOutput() string {
	return "ERR: Bad regex! pcre_compile() failed at position 1: missing )\n" +
		"If you meant to search for a literal string, run ag with -Q\n"
}

func TestRuleAgLiteral(t *testing.T) {
	resetTestState(t)
	rule := Get("ag_literal")
	if rule == nil {
		t.Fatal("ag_literal rule is not registered")
	}

	t.Run("match", func(t *testing.T) {
		assertMatch(t, rule, `ag \(`, agLiteralOutput())
	})

	t.Run("not match", func(t *testing.T) {
		assertNotMatch(t, rule, "ag foo", "")
	})

	t.Run("new command", func(t *testing.T) {
		assertNewCommands(t, rule, `ag \(`, agLiteralOutput(),
			[]string{`ag -Q \(`})
	})
}
