package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// glabUnknownCommandMatch matches the typo error that the GitLab CLI prints
// when a command or subcommand is misspelled and Cobra can suggest corrections.
func glabUnknownCommandMatch(command *types.Command) bool {
	output := command.Output
	return strings.Contains(output, `unknown command "`) &&
		strings.Contains(output, ` for "glab`) &&
		strings.Contains(output, "Did you mean this?")
}

// glabUnknownCommandNewCommand replaces the misspelled command with each
// suggestion of the GitLab CLI.
func glabUnknownCommandNewCommand(command *types.Command) ([]string, error) {
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
		"glab_unknown_command",
		utils.ForApp([]string{"glab"}, 0, glabUnknownCommandMatch),
		glabUnknownCommandNewCommand,
	))
}
