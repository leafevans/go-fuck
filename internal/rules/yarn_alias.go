package rules

import (
	"regexp"
	"strings"

	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// yarnAliasSuggestion matches the alias yarn suggests in its error output.
var yarnAliasSuggestion = regexp.MustCompile("Did you mean [`\"](?:yarn )?([^`\"]*)[`\"]")

// yarnAliasMatch reports whether yarn suggested a corrected alias.
func yarnAliasMatch(command *types.Command) bool {
	return strings.Contains(command.Output, "Did you mean")
}

// yarnAliasNewCommand replaces the unknown alias with the suggested one.
func yarnAliasNewCommand(command *types.Command) ([]string, error) {
	parts := command.ScriptParts()
	if len(parts) < 2 {
		return nil, nil
	}
	match := yarnAliasSuggestion.FindStringSubmatch(command.Output)
	if match == nil {
		return nil, nil
	}
	return []string{utils.ReplaceArgument(command.Script, parts[1], match[1])}, nil
}

func init() {
	register(types.NewRule("yarn_alias",
		utils.ForApp([]string{"yarn"}, 1, yarnAliasMatch),
		yarnAliasNewCommand))
}
