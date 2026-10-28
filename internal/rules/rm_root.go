package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/internal/specific"
	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// rmRootMatch reports whether rm refused to remove the root directory.
func rmRootMatch(command *types.Command) bool {
	parts := command.ScriptParts()
	if len(parts) == 0 || !utils.Contains(parts, "rm") || !utils.Contains(parts, "/") {
		return false
	}
	return !strings.Contains(command.Script, "--no-preserve-root") &&
		strings.Contains(command.Output, "--no-preserve-root")
}

// rmRootNewCommand allows removing the root directory.
func rmRootNewCommand(command *types.Command) ([]string, error) {
	return []string{command.Script + " --no-preserve-root"}, nil
}

func init() {
	rule := types.NewRule("rm_root",
		specific.SudoSupportMatch(rmRootMatch),
		specific.SudoSupportNew(rmRootNewCommand))
	// Python: enabled_by_default = False
	rule.EnabledByDefault = false
	register(rule)
}
