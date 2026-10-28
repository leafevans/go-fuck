package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/internal/specific"
	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// brewCaskDependencyMatch reports whether brew reported an unsatisfied cask
// dependency.
func brewCaskDependencyMatch(command *types.Command) bool {
	return utils.Contains(command.ScriptParts(), "install") &&
		strings.Contains(command.Output, "brew cask install")
}

// brewCaskDependencyCaskInstallLines returns the "brew cask install" commands
// suggested in the output.
func brewCaskDependencyCaskInstallLines(output string) []string {
	lines := []string{}
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "brew cask install") {
			lines = append(lines, line)
		}
	}
	return lines
}

// brewCaskDependencyNewCommand installs the missing casks and re-runs the
// command.
func brewCaskDependencyNewCommand(command *types.Command) ([]string, error) {
	lines := brewCaskDependencyCaskInstallLines(command.Output)
	if len(lines) == 0 {
		return nil, nil
	}
	caskScript := lines[0]
	if len(lines) > 1 {
		caskScript = strings.Join(lines, " && ")
	}
	return []string{caskScript + " && " + command.Script}, nil
}

func init() {
	rule := types.NewRule("brew_cask_dependency",
		utils.ForApp([]string{"brew"}, 0, brewCaskDependencyMatch),
		brewCaskDependencyNewCommand)
	rule.EnabledByDefault = specific.BrewAvailable()
	register(rule)
}
