package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// kubectlUnknownCommandMatch matches the typo error that kubectl prints when a
// command or subcommand is misspelled and Cobra can suggest corrections.
func kubectlUnknownCommandMatch(command *types.Command) bool {
	output := command.Output
	return strings.Contains(output, `unknown command "`) &&
		strings.Contains(output, ` for "kubectl`) &&
		strings.Contains(output, "Did you mean this?")
}

// kubectlUnknownCommandNewCommand replaces the misspelled command with each
// suggestion of kubectl.
func kubectlUnknownCommandNewCommand(command *types.Command) ([]string, error) {
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
		"kubectl_unknown_command",
		utils.ForApp([]string{"kubectl"}, 0, kubectlUnknownCommandMatch),
		kubectlUnknownCommandNewCommand,
	))
}
