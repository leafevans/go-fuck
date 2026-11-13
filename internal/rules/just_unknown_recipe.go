package rules

import (
	"regexp"

	"github.com/leafevans/go-fuck/v2/internal/types"
	"github.com/leafevans/go-fuck/v2/internal/utils"
)

var (
	// justUnknownRecipe matches the error line of just for an unknown recipe,
	// for example: error: justfile does not contain recipe `buidl`.
	justUnknownRecipe = regexp.MustCompile("justfile does not contain recipe `([^`]*)`")
	// justUnknownRecipeSuggestion matches the "Did you mean" line, for example:
	// Did you mean `build`?
	justUnknownRecipeSuggestion = regexp.MustCompile("Did you mean `([^`]*)`")
)

// justUnknownRecipeMatch reports whether just failed because the recipe does
// not exist and suggested a recipe that is similar to the unknown one.
func justUnknownRecipeMatch(command *types.Command) bool {
	if justUnknownRecipe.FindStringSubmatch(command.Output) == nil {
		return false
	}
	return justUnknownRecipeSuggestion.MatchString(command.Output)
}

// justUnknownRecipeNewCommand replaces the unknown recipe with the recipe that
// just suggests. A suggestion may also be an alias for another recipe; running
// the alias is valid, so it is used as is.
func justUnknownRecipeNewCommand(command *types.Command) ([]string, error) {
	broken := justUnknownRecipe.FindStringSubmatch(command.Output)
	if broken == nil {
		return nil, nil
	}
	suggestion := justUnknownRecipeSuggestion.FindStringSubmatch(command.Output)
	if suggestion == nil {
		return nil, nil
	}
	corrected := utils.ReplaceArgument(command.Script, broken[1], suggestion[1])
	if corrected == command.Script {
		return nil, nil
	}
	return []string{corrected}, nil
}

func init() {
	register(types.NewRule(
		"just_unknown_recipe",
		utils.ForApp([]string{"just"}, 0, justUnknownRecipeMatch),
		justUnknownRecipeNewCommand,
	))
}
