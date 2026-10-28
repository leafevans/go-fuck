package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/internal/specific"
	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// aptGetSearchMatch reports whether apt-get was called with the removed search
// command.
func aptGetSearchMatch(command *types.Command) bool {
	return strings.HasPrefix(command.Script, "apt-get search")
}

// aptGetSearchNewCommand replaces apt-get with apt-cache.
func aptGetSearchNewCommand(command *types.Command) ([]string, error) {
	script := command.Script
	if strings.HasPrefix(script, "apt-get") {
		script = "apt-cache" + script[len("apt-get"):]
	}
	return []string{script}, nil
}

func init() {
	rule := types.NewRule("apt_get_search",
		utils.ForApp([]string{"apt-get"}, 0, aptGetSearchMatch),
		aptGetSearchNewCommand)
	rule.EnabledByDefault = specific.AptAvailable()
	register(rule)
}
