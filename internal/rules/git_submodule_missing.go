package rules

import (
	"os/exec"
	"regexp"
	"strings"

	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// gitSubmoduleMissingPattern matches the pathspec error that git prints when a
// submodule path passed to "git submodule" does not exist.
var gitSubmoduleMissingPattern = regexp.MustCompile(`error: pathspec '([^']+)' did not match any file\(s\) known to git`)

// gitSubmodulePathsCommand lists the configured submodule paths. It is a
// variable so that tests do not depend on the repository that runs them.
var gitSubmodulePathsCommand = func() []string {
	output, err := exec.Command("git", "submodule", "status").Output()
	if err != nil {
		return nil
	}
	return parseSubmodulePaths(string(output))
}

// parseSubmodulePaths extracts the paths from "git submodule status" output.
func parseSubmodulePaths(output string) []string {
	paths := []string{}
	for _, line := range strings.Split(output, "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			paths = append(paths, fields[1])
		}
	}
	return paths
}

// gitSubmoduleMissingMatch reports whether "git submodule" failed because the
// path does not exist and another submodule path is close to it.
func gitSubmoduleMissingMatch(command *types.Command) bool {
	if !gitSubmoduleCommandUsesPath(command.Script) {
		return false
	}
	missing, ok := gitSubmoduleMissingName(command.Output)
	if !ok || missing == "" {
		return false
	}
	closest := utils.GetClosest(missing, gitSubmodulePathsCommand(), 0.6, false)
	return closest != "" && closest != missing
}

// gitSubmoduleMissingNewCommand replaces the missing submodule path with the
// closest configured submodule.
func gitSubmoduleMissingNewCommand(command *types.Command) ([]string, error) {
	if !gitSubmoduleCommandUsesPath(command.Script) {
		return nil, nil
	}
	missing, ok := gitSubmoduleMissingName(command.Output)
	if !ok || missing == "" {
		return nil, nil
	}
	closest := utils.GetClosest(missing, gitSubmodulePathsCommand(), 0.6, false)
	if closest == "" || closest == missing {
		return nil, nil
	}
	corrected := strings.Replace(command.Script, missing, closest, 1)
	if corrected == command.Script {
		return nil, nil
	}
	return []string{corrected}, nil
}

// gitSubmoduleCommandUsesPath reports whether script is a "git submodule"
// subcommand that takes an existing submodule path.
func gitSubmoduleCommandUsesPath(script string) bool {
	return strings.Contains(script, "submodule update ") ||
		strings.Contains(script, "submodule status ")
}

// gitSubmoduleMissingName returns the missing path from a git error.
func gitSubmoduleMissingName(output string) (string, bool) {
	if match := gitSubmoduleMissingPattern.FindStringSubmatch(output); match != nil {
		return match[1], true
	}
	return "", false
}

func init() {
	register(types.NewRule("git_submodule_missing",
		utils.ForApp([]string{"git"}, 0, gitSubmoduleMissingMatch),
		gitSubmoduleMissingNewCommand))
}
