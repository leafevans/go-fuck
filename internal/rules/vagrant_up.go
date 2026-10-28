package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/internal/shells"
	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// vagrantUpMatch reports whether vagrant tells the user to start the machine
// first.
func vagrantUpMatch(command *types.Command) bool {
	return strings.Contains(strings.ToLower(command.Output), "run `vagrant up`")
}

// vagrantUpNewCommand starts the machine and retries the command.
func vagrantUpNewCommand(command *types.Command) ([]string, error) {
	parts := command.ScriptParts()
	startAllInstances := shells.Current().And("vagrant up", command.Script)

	if len(parts) < 3 {
		return []string{startAllInstances}, nil
	}
	machine := parts[2]
	return []string{
		shells.Current().And("vagrant up "+machine, command.Script),
		startAllInstances,
	}, nil
}

func init() {
	register(types.NewRule("vagrant_up",
		utils.ForApp([]string{"vagrant"}, 0, vagrantUpMatch),
		vagrantUpNewCommand))
}
