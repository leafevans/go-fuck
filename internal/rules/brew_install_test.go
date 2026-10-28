package rules

import (
	"strings"
	"testing"
)

const brewInstallNoAvailableFormulaOne = `Warning: No available formula with the name "giss". Did you mean gist?`
const brewInstallNoAvailableFormulaTwo = `Warning: No available formula with the name "elasticserar". Did you mean elasticsearch or elasticsearch@6?`
const brewInstallNoAvailableFormulaThree = `Warning: No available formula with the name "gitt". Did you mean git, gitg or gist?`
const brewInstallNoArgument = `Install a formula or cask. Additional options specific to a formula may be`
const brewInstallAlreadyInstalled = `Warning: git-2.3.5 already installed`

func TestRuleBrewInstall(t *testing.T) {
	resetTestState(t)
	rule := Get("brew_install")
	if rule == nil {
		t.Fatal("brew_install rule is not registered")
	}

	suggestionCases := []struct {
		name        string
		suggestions string
		want        []string
	}{
		{"single", "one", []string{"one"}},
		{"or", "one or two", []string{"one", "two"}},
		{"comma_and_or", "one, two or three", []string{"one", "two", "three"}},
	}
	for _, testCase := range suggestionCases {
		t.Run("suggestions_"+testCase.name, func(t *testing.T) {
			got := brewInstallSuggestions(testCase.suggestions)
			if strings.Join(got, ",") != strings.Join(testCase.want, ",") {
				t.Errorf("_get_suggestions(%q) = %q, want %q",
					testCase.suggestions, got, testCase.want)
			}
		})
	}

	assertMatch(t, rule, "brew install giss", brewInstallNoAvailableFormulaOne)
	assertMatch(t, rule, "brew install elasticserar", brewInstallNoAvailableFormulaTwo)
	assertMatch(t, rule, "brew install gitt", brewInstallNoAvailableFormulaThree)
	assertNotMatch(t, rule, "brew install git", brewInstallAlreadyInstalled)
	assertNotMatch(t, rule, "brew install", brewInstallNoArgument)

	assertNewCommands(t, rule, "brew install giss", brewInstallNoAvailableFormulaOne,
		[]string{"brew install gist"})
	assertNewCommands(t, rule, "brew install elasticsear", brewInstallNoAvailableFormulaTwo,
		[]string{"brew install elasticsearch", "brew install elasticsearch@6"})
	assertNewCommands(t, rule, "brew install gitt", brewInstallNoAvailableFormulaThree,
		[]string{"brew install git", "brew install gitg", "brew install gist"})
	assertNewCommands(t, rule, "brew install aa", brewInstallNoAvailableFormulaOne,
		[]string{"brew install gist"})
}
