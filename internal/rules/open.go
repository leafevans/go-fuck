package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/internal/shells"
	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// openURLMarkers lists the fragments that mark a command argument as a URL.
var openURLMarkers = []string{
	".com", ".edu", ".info", ".io", ".ly", ".me", ".net", ".org", ".se",
	"www.",
}

// openIsArgURL reports whether the command opens a URL.
func openIsArgURL(command *types.Command) bool {
	for _, marker := range openURLMarkers {
		if strings.Contains(command.Script, marker) {
			return true
		}
	}
	return false
}

// openMatch reports whether open was called on a URL or on a missing file.
func openMatch(command *types.Command) bool {
	output := strings.TrimSpace(command.Output)
	return openIsArgURL(command) ||
		strings.HasPrefix(output, "The file ") &&
			strings.HasSuffix(output, " does not exist.")
}

// openNewCommand opens the URL or creates the missing file first.
func openNewCommand(command *types.Command) ([]string, error) {
	output := strings.TrimSpace(command.Output)
	if openIsArgURL(command) {
		return []string{strings.ReplaceAll(command.Script, "open ", "open http://")}, nil
	}
	if strings.HasPrefix(output, "The file ") && strings.HasSuffix(output, " does not exist.") {
		fields := strings.SplitN(command.Script, " ", 2)
		if len(fields) < 2 {
			return nil, nil
		}
		argument := fields[1]
		return []string{
			shells.Current().And("touch "+argument, command.Script),
			shells.Current().And("mkdir "+argument, command.Script),
		}, nil
	}
	return nil, nil
}

func init() {
	register(types.NewRule("open",
		utils.ForApp([]string{"open", "xdg-open", "gnome-open", "kde-open"}, 0, openMatch),
		openNewCommand))
}
