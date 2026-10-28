package rules

import (
	"regexp"
	"strings"

	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// goInstallSuggestion matches the command that the Go toolchain suggests for
// go install outside of a module.
var goInstallSuggestion = regexp.MustCompile(`Try '(go install [^']*)' to install the latest version`)

// goInstallVersionMatch matches the error that the Go toolchain prints when
// go install is used outside a module without a version.
func goInstallVersionMatch(command *types.Command) bool {
	output := command.Output
	return strings.Contains(output, "'go install' requires a version") &&
		strings.Contains(output, "to install the latest version")
}

// goInstallVersionNewCommand returns the command that the Go toolchain
// suggests, which appends the missing version.
func goInstallVersionNewCommand(command *types.Command) ([]string, error) {
	match := goInstallSuggestion.FindStringSubmatch(command.Output)
	if match == nil {
		return nil, nil
	}
	return []string{match[1]}, nil
}

func init() {
	register(types.NewRule(
		"go_install_version",
		utils.ForApp([]string{"go"}, 0, goInstallVersionMatch),
		goInstallVersionNewCommand,
	))
}
