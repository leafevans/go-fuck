package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/v3/internal/types"
	"github.com/leafevans/go-fuck/v3/internal/utils"
)

// gitUnknownRemotePatterns match the errors that git prints when a remote of a
// "git remote" subcommand does not exist.
var gitUnknownRemotePatterns = []*lazyRegexp{
	newRegexp(`(?m)^(?:error|fatal): No such remote: '([^']+)'$`),
	newRegexp(`(?m)^(?:error|fatal): No such remote '([^']+)'$`),
	newRegexp(`(?m)^fatal: '([^']+)' does not appear to be a git repository$`),
}

// gitRemoteNamesCommand lists the configured remotes. It is a variable so that
// tests do not depend on the repository that runs them.
var gitRemoteNamesCommand = func(script string) []string {
	output, err := gitRepositoryOutput(script, "remote", "remote")
	if err != nil {
		return nil
	}
	return strings.Fields(string(output))
}

// gitUnknownRemoteMatch reports whether a "git remote" subcommand failed
// because the remote name does not exist and another configured remote is
// close to it.
func gitUnknownRemoteMatch(command *types.Command) bool {
	if !gitRemoteCommandUsesName(command.Script) {
		return false
	}
	missing, ok := gitUnknownRemoteName(command.Output)
	if !ok || missing == "" {
		return false
	}
	closest := utils.GetClosest(missing, gitRemoteNamesCommand(command.Script), 0.6, false)
	return closest != "" && closest != missing
}

// gitUnknownRemoteNewCommand replaces the unknown remote with the closest
// configured remote.
func gitUnknownRemoteNewCommand(command *types.Command) ([]string, error) {
	if !gitRemoteCommandUsesName(command.Script) {
		return nil, nil
	}
	missing, ok := gitUnknownRemoteName(command.Output)
	if !ok || missing == "" {
		return nil, nil
	}
	closest := utils.GetClosest(missing, gitRemoteNamesCommand(command.Script), 0.6, false)
	if closest == "" || closest == missing {
		return nil, nil
	}
	corrected, ok := gitReplaceArgument(command.Script, missing, closest, false)
	if !ok {
		return nil, nil
	}
	return []string{corrected}, nil
}

// gitRemoteCommandUsesName reports whether script is a "git remote"
// subcommand that takes an existing remote name. "add" reports a different
// error.
func gitRemoteCommandUsesName(script string) bool {
	return gitHasOperation(script, "remote", "remove", "rm", "rename", "get-url", "set-url", "show", "prune")
}

// gitUnknownRemoteName returns the unknown remote name from a git error.
func gitUnknownRemoteName(output string) (string, bool) {
	for _, pattern := range gitUnknownRemotePatterns {
		if match := pattern.FindStringSubmatch(output); match != nil {
			return match[1], true
		}
	}
	return "", false
}

func init() {
	register(types.NewRule("git_unknown_remote",
		utils.ForApp([]string{"git"}, 0, gitUnknownRemoteMatch),
		gitUnknownRemoteNewCommand))
}
