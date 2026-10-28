package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/internal/specific"
	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// aptUpgradeMatch reports whether apt printed a list of upgradable packages.
func aptUpgradeMatch(command *types.Command) bool {
	if command.Script != "apt list --upgradable" {
		return false
	}
	return len(strings.Split(strings.TrimSpace(command.Output), "\n")) > 1
}

// aptUpgradeNewCommand upgrades the installed packages.
func aptUpgradeNewCommand(command *types.Command) ([]string, error) {
	return []string{"apt upgrade"}, nil
}

func init() {
	rule := types.NewRule("apt_upgrade",
		specific.SudoSupportMatch(utils.ForApp([]string{"apt"}, 0, aptUpgradeMatch)),
		specific.SudoSupportNew(aptUpgradeNewCommand))
	rule.EnabledByDefault = specific.AptAvailable()
	register(rule)
}
