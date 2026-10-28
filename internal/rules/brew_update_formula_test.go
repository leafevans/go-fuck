package rules

import "testing"

const brewUpdateFormulaOutput = "Error: This command updates brew itself, and does not take formula" +
	" names.\nUse `brew upgrade thefuck`."

func TestRuleBrewUpdateFormula(t *testing.T) {
	resetTestState(t)
	rule := Get("brew_update_formula")
	if rule == nil {
		t.Fatal("brew_update_formula rule is not registered")
	}

	assertMatch(t, rule, "brew update thefuck", brewUpdateFormulaOutput)
	assertNotMatch(t, rule, "brew upgrade foo", "")
	assertNotMatch(t, rule, "brew update", "")

	assertNewCommands(t, rule, "brew update foo", brewUpdateFormulaOutput,
		[]string{"brew upgrade foo"})
	assertNewCommands(t, rule, "brew update bar zap", brewUpdateFormulaOutput,
		[]string{"brew upgrade bar zap"})
}
