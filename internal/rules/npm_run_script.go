package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/internal/specific"
	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// npmRunScriptUsage is the first line of the npm help output.
const npmRunScriptUsage = "Usage: npm <command>"

// npmRunScriptRunPartPrefix is the prefix of the npm arguments that run a
// script ("run", "ru", "run-script", ...).
const npmRunScriptRunPartPrefix = "ru"

// npmRunScriptMatch reports whether npm printed its help because the script
// was passed without "run".
func npmRunScriptMatch(command *types.Command) bool {
	if !strings.Contains(command.Output, npmRunScriptUsage) {
		return false
	}
	parts := command.ScriptParts()
	if len(parts) < 2 {
		return false
	}
	for _, part := range parts {
		if strings.HasPrefix(part, npmRunScriptRunPartPrefix) {
			return false
		}
	}
	return utils.Contains(specific.NpmGetScripts(), parts[1])
}

// npmRunScriptNewCommand inserts the run-script argument.
func npmRunScriptNewCommand(command *types.Command) ([]string, error) {
	parts := command.ScriptParts()
	if len(parts) < 2 {
		return nil, nil
	}
	fixed := make([]string, 0, len(parts)+1)
	fixed = append(fixed, parts[0], "run-script")
	fixed = append(fixed, parts[1:]...)
	return []string{strings.Join(fixed, " ")}, nil
}

func init() {
	// Python: enabled_by_default = npm_available
	rule := types.NewRule("npm_run_script",
		utils.ForApp([]string{"npm"}, 0, npmRunScriptMatch),
		npmRunScriptNewCommand)
	rule.EnabledByDefault = specific.NpmAvailable()
	register(rule)
}
