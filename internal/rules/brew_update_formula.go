package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// brewUpdateFormulaMatch reports whether brew was asked to update a formula.
func brewUpdateFormulaMatch(command *types.Command) bool {
	return strings.Contains(command.Script, "update") &&
		strings.Contains(command.Output, "Error: This command updates brew itself") &&
		strings.Contains(command.Output, "Use `brew upgrade")
}

// brewUpdateFormulaNewCommand upgrades the given formulas instead.
func brewUpdateFormulaNewCommand(command *types.Command) ([]string, error) {
	return []string{strings.ReplaceAll(command.Script, "update", "upgrade")}, nil
}

func init() {
	register(types.NewRule("brew_update_formula",
		utils.ForApp([]string{"brew"}, 2, brewUpdateFormulaMatch),
		brewUpdateFormulaNewCommand))
}
