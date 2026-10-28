package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/internal/types"
)

// removeTrailingCedillaCedilla is the cedilla that is accidentally typed at
// the end of a command.
const removeTrailingCedillaCedilla = "\u00e7"

// removeTrailingCedillaMatch reports whether the command ends with a cedilla.
func removeTrailingCedillaMatch(command *types.Command) bool {
	return strings.HasSuffix(command.Script, removeTrailingCedillaCedilla)
}

// removeTrailingCedillaNewCommand removes the trailing cedilla.
func removeTrailingCedillaNewCommand(command *types.Command) ([]string, error) {
	return []string{strings.TrimSuffix(command.Script, removeTrailingCedillaCedilla)}, nil
}

func init() {
	register(types.NewRule("remove_trailing_cedilla",
		removeTrailingCedillaMatch,
		removeTrailingCedillaNewCommand))
}
