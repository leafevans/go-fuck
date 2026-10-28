package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// brewUninstallMatch reports whether brew could not uninstall all versions of
// a formula.
func brewUninstallMatch(command *types.Command) bool {
	action := command.ScriptParts()[1]
	return (action == "uninstall" || action == "rm" || action == "remove") &&
		strings.Contains(command.Output, "brew uninstall --force")
}

// brewUninstallNewCommand removes all installed versions of the formula.
func brewUninstallNewCommand(command *types.Command) ([]string, error) {
	parts := command.ScriptParts()
	newParts := append([]string{}, parts[:2]...)
	newParts = append(newParts, "--force")
	newParts = append(newParts, parts[2:]...)
	return []string{strings.Join(newParts, " ")}, nil
}

func init() {
	register(types.NewRule("brew_uninstall",
		utils.ForApp([]string{"brew"}, 2, brewUninstallMatch),
		brewUninstallNewCommand))
}
