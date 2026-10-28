package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/internal/specific"
	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// aptListUpgradableMatch reports whether apt printed the command that lists the
// upgradable packages.
func aptListUpgradableMatch(command *types.Command) bool {
	return strings.Contains(command.Output, "apt list --upgradable")
}

// aptListUpgradableNewCommand lists the upgradable packages.
func aptListUpgradableNewCommand(command *types.Command) ([]string, error) {
	return []string{"apt list --upgradable"}, nil
}

func init() {
	rule := types.NewRule("apt_list_upgradable",
		specific.SudoSupportMatch(utils.ForApp([]string{"apt"}, 0, aptListUpgradableMatch)),
		specific.SudoSupportNew(aptListUpgradableNewCommand))
	rule.EnabledByDefault = specific.AptAvailable()
	register(rule)
}
