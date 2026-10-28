package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/internal/specific"
	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// chocoInstallExcludedParts lists the script parts that are not package names.
var chocoInstallExcludedParts = []string{"choco", "cinst", "install"}

// chocoInstallMatch reports whether Chocolatey could not find a package.
func chocoInstallMatch(command *types.Command) bool {
	if !strings.HasPrefix(command.Script, "choco install") &&
		!utils.Contains(command.ScriptParts(), "cinst") {
		return false
	}
	return strings.Contains(command.Output, "Installing the following packages")
}

// chocoInstallNewCommand appends the ".install" suffix that forces the
// Chocolatey install package instead of an executable package.
func chocoInstallNewCommand(command *types.Command) ([]string, error) {
	for _, part := range command.ScriptParts() {
		if utils.Contains(chocoInstallExcludedParts, part) ||
			strings.HasPrefix(part, "-") ||
			strings.Contains(part, "=") ||
			strings.Contains(part, "/") {
			continue
		}
		// Need an exact match because "chocolatey" is a valid package name.
		return []string{strings.ReplaceAll(command.Script, part, part+".install")}, nil
	}
	return nil, nil
}

func init() {
	rule := types.NewRule("choco_install",
		utils.ForApp([]string{"choco", "cinst"}, 0, chocoInstallMatch),
		chocoInstallNewCommand)
	rule.EnabledByDefault = specific.ChocoAvailable()
	register(rule)
}
