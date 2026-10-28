package rules

import (
	"regexp"
	"strings"

	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// brewReinstallWarningRegexp matches the "already installed" warning of brew.
//
// Deviation: the original Python rule uses a negative lookahead here. RE2 has
// no lookarounds, so the captured formula is checked for an embedded "is "
// instead.
var brewReinstallWarningRegexp = regexp.MustCompile(
	`Warning: (.+) is already installed and up-to-date`)

// brewReinstallMessageRegexp matches the reinstall suggestion of brew.
//
// Deviation: the original Python rule uses a negative lookahead here. RE2 has
// no lookarounds, so the captured formula is checked for an embedded ", "
// instead.
var brewReinstallMessageRegexp = regexp.MustCompile(
	"To reinstall (.+?), run `brew reinstall [^`]+`")

// brewReinstallHasWarning reports whether the output contains the warning.
func brewReinstallHasWarning(output string) bool {
	for _, match := range brewReinstallWarningRegexp.FindAllStringSubmatch(output, -1) {
		if !strings.Contains(match[1], "is ") {
			return true
		}
	}
	return false
}

// brewReinstallHasMessage reports whether the output contains the reinstall
// suggestion.
func brewReinstallHasMessage(output string) bool {
	for _, match := range brewReinstallMessageRegexp.FindAllStringSubmatch(output, -1) {
		if !strings.Contains(match[1], ", ") {
			return true
		}
	}
	return false
}

// brewReinstallMatch reports whether brew reported that the latest version is
// already installed.
func brewReinstallMatch(command *types.Command) bool {
	return strings.Contains(command.Script, "install") &&
		brewReinstallHasWarning(command.Output) &&
		brewReinstallHasMessage(command.Output)
}

// brewReinstallNewCommand reinstalls the already installed formula.
func brewReinstallNewCommand(command *types.Command) ([]string, error) {
	return []string{strings.ReplaceAll(command.Script, "install", "reinstall")}, nil
}

func init() {
	register(types.NewRule("brew_reinstall",
		utils.ForApp([]string{"brew"}, 2, brewReinstallMatch),
		brewReinstallNewCommand))
}
