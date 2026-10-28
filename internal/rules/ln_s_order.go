package rules

import (
	"os"
	"strings"

	"github.com/leafevans/go-fuck/internal/specific"
	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// lnSOrderPathExists reports whether a command argument exists. It is a
// variable so that tests can replace the file system lookup.
var lnSOrderPathExists = func(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// lnSOrderDestination returns the destination argument of a misordered ln
// command. When the arguments are swapped, the first existing argument is the
// destination.
func lnSOrderDestination(parts []string) string {
	for _, part := range parts {
		if part == "ln" || part == "-s" || part == "--symbolic" {
			continue
		}
		if lnSOrderPathExists(part) {
			return part
		}
	}
	return ""
}

// lnSOrderMatch reports whether ln failed because the symbolic link arguments
// are swapped.
func lnSOrderMatch(command *types.Command) bool {
	parts := command.ScriptParts()
	if len(parts) == 0 || parts[0] != "ln" {
		return false
	}
	if !utils.Contains(parts, "-s") && !utils.Contains(parts, "--symbolic") {
		return false
	}
	return strings.Contains(command.Output, "File exists") &&
		lnSOrderDestination(parts) != ""
}

// lnSOrderNewCommand moves the destination to the end of the command.
func lnSOrderNewCommand(command *types.Command) ([]string, error) {
	parts := command.ScriptParts()
	destination := lnSOrderDestination(parts)
	if destination == "" {
		return nil, nil
	}

	moved := append([]string{}, parts...)
	if index := utils.Index(moved, destination); index >= 0 {
		moved = append(moved[:index], moved[index+1:]...)
	}
	moved = append(moved, destination)
	return []string{strings.Join(moved, " ")}, nil
}

func init() {
	register(types.NewRule("ln_s_order",
		specific.SudoSupportMatch(lnSOrderMatch),
		specific.SudoSupportNew(lnSOrderNewCommand)))
}
