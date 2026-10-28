package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/internal/types"
)

// unsudoPatterns lists the error messages that indicate the command must not
// be run as root.
var unsudoPatterns = []string{
	"you cannot perform this operation as root",
}

// unsudoMatch reports whether the command failed because it was run as root.
func unsudoMatch(command *types.Command) bool {
	parts := command.ScriptParts()
	if len(parts) > 0 && parts[0] != "sudo" {
		return false
	}

	output := strings.ToLower(command.Output)
	for _, pattern := range unsudoPatterns {
		if strings.Contains(output, pattern) {
			return true
		}
	}
	return false
}

// unsudoNewCommand removes the sudo prefix.
func unsudoNewCommand(command *types.Command) ([]string, error) {
	parts := command.ScriptParts()
	if len(parts) > 1 {
		parts = parts[1:]
	} else {
		parts = nil
	}
	return []string{strings.Join(parts, " ")}, nil
}

func init() {
	register(types.NewRule("unsudo", unsudoMatch, unsudoNewCommand))
}
