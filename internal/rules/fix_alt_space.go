package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/internal/specific"
	"github.com/leafevans/go-fuck/internal/types"
)

// fixAltSpaceAltSpace is the non-breaking space that is typed instead of a
// regular space when the Alt key is held down for too long.
const fixAltSpaceAltSpace = "\u00a0"

// fixAltSpaceMatch reports whether the command contains a non-breaking space
// and a command was not found.
func fixAltSpaceMatch(command *types.Command) bool {
	return strings.Contains(strings.ToLower(command.Output), "command not found") &&
		strings.Contains(command.Script, fixAltSpaceAltSpace)
}

// fixAltSpaceNewCommand replaces the non-breaking spaces with regular spaces.
func fixAltSpaceNewCommand(command *types.Command) ([]string, error) {
	return []string{strings.ReplaceAll(command.Script, fixAltSpaceAltSpace, " ")}, nil
}

func init() {
	register(types.NewRule("fix_alt_space",
		specific.SudoSupportMatch(fixAltSpaceMatch),
		specific.SudoSupportNew(fixAltSpaceNewCommand)))
}
