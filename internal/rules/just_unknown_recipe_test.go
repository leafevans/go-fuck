package rules

import "testing"

// justUnknownRecipeOutput is the output of just for an unknown recipe.
func justUnknownRecipeOutput() string {
	return "error: justfile does not contain recipe `buidl`\n" +
		"Did you mean `build`?\n"
}

func TestRuleJustUnknownRecipe(t *testing.T) {
	resetTestState(t)
	rule := Get("just_unknown_recipe")
	if rule == nil {
		t.Fatal("just_unknown_recipe rule is not registered")
	}

	assertMatch(t, rule, "just buidl", justUnknownRecipeOutput())
	assertNotMatch(t, rule, "cargo buidl", justUnknownRecipeOutput())
	// Without a suggestion there is nothing to correct.
	assertNotMatch(t, rule, "just buidl", "error: justfile does not contain recipe `buidl`\n")

	assertNewCommands(t, rule, "just buidl", justUnknownRecipeOutput(), []string{"just build"})
	assertNewCommands(t, rule, "just buidl --release", justUnknownRecipeOutput(), []string{"just build --release"})
}

func TestRuleJustUnknownRecipeAlias(t *testing.T) {
	resetTestState(t)
	rule := Get("just_unknown_recipe")

	output := "error: justfile does not contain recipe `buidl`\n" +
		"Did you mean `bld`, an alias for `build`?\n"
	assertNewCommands(t, rule, "just buidl", output, []string{"just bld"})
}
