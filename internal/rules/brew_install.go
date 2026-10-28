package rules

import (
	"regexp"
	"strings"

	"github.com/leafevans/go-fuck/internal/specific"
	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// brewInstallSuggestionRegexp matches the formula suggestions of the brew
// warning.
var brewInstallSuggestionRegexp = regexp.MustCompile(
	`Warning: No available formula with the name "[^"]+"\. Did you mean (.+)\?`)

// brewInstallSuggestions splits the suggestion list of the brew warning.
func brewInstallSuggestions(suggestions string) []string {
	return strings.Split(strings.ReplaceAll(suggestions, " or ", ", "), ", ")
}

// brewInstallMatch reports whether brew suggested formula names.
func brewInstallMatch(command *types.Command) bool {
	return strings.Contains(command.Script, "install") &&
		strings.Contains(command.Output, "No available formula") &&
		strings.Contains(command.Output, "Did you mean")
}

// brewInstallNewCommand suggests install commands for the closest formulas.
func brewInstallNewCommand(command *types.Command) ([]string, error) {
	match := brewInstallSuggestionRegexp.FindStringSubmatch(command.Output)
	if match == nil {
		return nil, nil
	}
	results := []string{}
	for _, suggestion := range brewInstallSuggestions(match[1]) {
		results = append(results, "brew install "+suggestion)
	}
	return results, nil
}

func init() {
	rule := types.NewRule("brew_install",
		utils.ForApp([]string{"brew"}, 2, brewInstallMatch),
		brewInstallNewCommand)
	rule.EnabledByDefault = specific.BrewAvailable()
	register(rule)
}
