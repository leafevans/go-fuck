package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// mvnNoCommandMatch reports whether maven was called without a goal.
func mvnNoCommandMatch(command *types.Command) bool {
	return strings.Contains(command.Output, "No goals have been specified for this build")
}

// mvnNoCommandNewCommand suggests the most common maven goals.
func mvnNoCommandNewCommand(command *types.Command) ([]string, error) {
	return []string{
		command.Script + " clean package",
		command.Script + " clean install",
	}, nil
}

func init() {
	register(types.NewRule("mvn_no_command",
		utils.ForApp([]string{"mvn"}, 0, mvnNoCommandMatch),
		mvnNoCommandNewCommand))
}
