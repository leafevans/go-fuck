package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/v4/internal/types"
	"github.com/leafevans/go-fuck/v4/internal/utils"
)

// yarnAliasSuggestion matches the alias yarn suggests in its error output.
var yarnAliasSuggestion = newRegexp("Did you mean [`\"](?:yarn )?([^`\"]*)[`\"]")

// yarnAliasMatch reports whether yarn suggested a corrected alias.
func yarnAliasMatch(command *types.Command) bool {
	return strings.Contains(command.Output, "Did you mean")
}

// yarnAliasNewCommand replaces the unknown alias with the suggested one.
func yarnAliasNewCommand(command *types.Command) ([]string, error) {
	words := simpleCommandWords(command.Script)
	if len(words) < 2 || words[1].text == "" || strings.HasPrefix(words[1].text, "-") {
		return nil, nil
	}
	match := yarnAliasSuggestion.FindStringSubmatch(command.Output)
	if match == nil {
		return nil, nil
	}
	return replaceCommandWord(command.Script, words[1], "", match[1]), nil
}

func init() {
	register(types.NewRule("yarn_alias",
		utils.ForApp([]string{"yarn"}, 1, yarnAliasMatch),
		yarnAliasNewCommand))
}
