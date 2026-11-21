package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/v2/internal/types"
	"github.com/leafevans/go-fuck/v2/internal/utils"
)

// gitSubmoduleMissingPattern matches the pathspec error that git prints when a
// submodule path passed to "git submodule" does not exist.
var gitSubmoduleMissingPattern = newRegexp(`error: pathspec '([^']+)' did not match any file\(s\) known to git`)

// gitSubmodulePathsCommand lists the configured submodule paths. It is a
// variable so that tests do not depend on the repository that runs them.
var gitSubmodulePathsCommand = func(script string) []string {
	output, err := gitRepositoryOutput(script, "submodule", "ls-files", "--stage", "-z")
	if err != nil {
		return nil
	}
	return parseSubmodulePaths(string(output))
}

// parseSubmodulePaths extracts gitlinks from the NUL-delimited index records.
func parseSubmodulePaths(output string) []string {
	paths := []string{}
	for record := range strings.SplitSeq(output, "\x00") {
		meta, path, ok := strings.Cut(record, "\t")
		if ok && strings.HasPrefix(meta, "160000 ") && strings.HasSuffix(meta, " 0") && path != "" {
			paths = append(paths, path)
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
	closest := utils.GetClosest(missing, gitSubmodulePathsCommand(command.Script), 0.6, false)
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
	closest := utils.GetClosest(missing, gitSubmodulePathsCommand(command.Script), 0.6, false)
	if closest == "" || closest == missing {
		return nil, nil
	}
	corrected, ok := gitReplaceArgument(command.Script, missing, closest, true)
	if !ok {
		return nil, nil
	}
	return []string{corrected}, nil
}

// gitSubmoduleCommandUsesPath reports whether script is a "git submodule"
// subcommand that takes an existing submodule path.
func gitSubmoduleCommandUsesPath(script string) bool {
	return gitHasOperation(script, "submodule", "update", "status")
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
