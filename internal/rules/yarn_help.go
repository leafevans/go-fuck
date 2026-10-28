package rules

import (
	"regexp"
	"strings"

	"github.com/leafevans/go-fuck/internal/system"
	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// yarnHelpDocumentation matches the documentation URL printed by yarn help.
var yarnHelpDocumentation = regexp.MustCompile(`Visit ([^ ]*) for documentation about this command\.`)

// yarnHelpMatch reports whether yarn asked to visit the command documentation.
func yarnHelpMatch(command *types.Command) bool {
	parts := command.ScriptParts()
	if len(parts) < 2 {
		return false
	}
	return parts[1] == "help" &&
		strings.Contains(command.Output, "for documentation about this command.")
}

// yarnHelpNewCommand opens the documentation URL in the browser.
func yarnHelpNewCommand(command *types.Command) ([]string, error) {
	match := yarnHelpDocumentation.FindStringSubmatch(command.Output)
	if match == nil {
		return nil, nil
	}
	return []string{system.OpenCommand(match[1])}, nil
}

func init() {
	register(types.NewRule("yarn_help",
		utils.ForApp([]string{"yarn"}, 2, yarnHelpMatch),
		yarnHelpNewCommand))
}
