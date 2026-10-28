package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/internal/specific"
	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// pipInstallMatch reports whether pip install failed with a permission error.
func pipInstallMatch(command *types.Command) bool {
	return strings.Contains(command.Script, "pip install") &&
		strings.Contains(command.Output, "Permission denied")
}

// pipInstallNewCommand retries the installation with --user and falls back to
// sudo when --user was already used.
func pipInstallNewCommand(command *types.Command) ([]string, error) {
	if !strings.Contains(command.Script, "--user") {
		return []string{strings.ReplaceAll(command.Script, " install ", " install --user ")}, nil
	}
	return []string{"sudo " + strings.ReplaceAll(command.Script, " --user", "")}, nil
}

func init() {
	register(types.NewRule("pip_install",
		specific.SudoSupportMatch(utils.ForApp([]string{"pip"}, 0, pipInstallMatch)),
		pipInstallNewCommand))
}
