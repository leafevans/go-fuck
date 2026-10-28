package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/internal/specific"
	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// systemctlMatch reports whether systemctl was called with the service name
// and the operation in the wrong order.
func systemctlMatch(command *types.Command) bool {
	parts := command.ScriptParts()
	if len(parts) == 0 || !strings.Contains(command.Output, "Unknown operation '") {
		return false
	}
	index := utils.Index(parts, "systemctl")
	if index < 0 {
		return false
	}
	return len(parts)-index == 3
}

// systemctlNewCommand swaps the service name and the operation.
func systemctlNewCommand(command *types.Command) ([]string, error) {
	parts := append([]string{}, command.ScriptParts()...)
	if len(parts) < 2 {
		return nil, nil
	}
	parts[len(parts)-1], parts[len(parts)-2] = parts[len(parts)-2], parts[len(parts)-1]
	return []string{strings.Join(parts, " ")}, nil
}

func init() {
	register(types.NewRule("systemctl",
		specific.SudoSupportMatch(utils.ForApp([]string{"systemctl"}, 0, systemctlMatch)),
		specific.SudoSupportNew(systemctlNewCommand)))
}
