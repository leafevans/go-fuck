package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/v4/internal/types"
	"github.com/leafevans/go-fuck/v4/internal/utils"
)

// flyctlUnknownCommandMatch reports whether flyctl, whose root command is
// named "fly", printed a Cobra unknown command error. Nested groups include
// their path in the error, for example 'for "fly mpg"'.
func flyctlUnknownCommandMatch(command *types.Command) bool {
	output := command.Output
	return strings.Contains(output, `unknown command "`) &&
		strings.Contains(output, ` for "fly`) &&
		strings.Contains(output, "Did you mean this?")
}

// flyctlUnknownCommandNewCommand replaces the unknown command with the
// suggestions that Cobra printed.
func flyctlUnknownCommandNewCommand(command *types.Command) ([]string, error) {
	return cobraUnknownCommandNewCommand(command)
}

func init() {
	register(types.NewRule(
		"flyctl_unknown_command",
		utils.ForApp([]string{"flyctl", "fly"}, 0, flyctlUnknownCommandMatch),
		flyctlUnknownCommandNewCommand,
	))
}
