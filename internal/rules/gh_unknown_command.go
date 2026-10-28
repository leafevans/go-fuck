package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// ghUnknownCommandMatch matches the typo error that the GitHub CLI prints when
// a command or subcommand is misspelled and Cobra can suggest corrections.
func ghUnknownCommandMatch(command *types.Command) bool {
	output := command.Output
	return strings.Contains(output, `unknown command "`) &&
		strings.Contains(output, ` for "gh`) &&
		strings.Contains(output, "Did you mean this?")
}

// ghUnknownCommandNewCommand replaces the misspelled command with each
// suggestion of the GitHub CLI.
func ghUnknownCommandNewCommand(command *types.Command) ([]string, error) {
	broken, suggestions := cobraSuggestions(command.Output)
	if broken == "" {
		return nil, nil
	}

	corrected := make([]string, 0, len(suggestions))
	for _, suggestion := range suggestions {
		corrected = append(corrected, utils.ReplaceArgument(command.Script, broken, suggestion))
	}
	return corrected, nil
}

func init() {
	register(types.NewRule(
		"gh_unknown_command",
		utils.ForApp([]string{"gh"}, 0, ghUnknownCommandMatch),
		ghUnknownCommandNewCommand,
	))
}
