package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/v2/internal/specific"
	"github.com/leafevans/go-fuck/v2/internal/types"
	"github.com/leafevans/go-fuck/v2/internal/utils"
)

// gitPullDivergentError matches the fatal error that git prints when a pull
// has to merge divergent branches and no reconciliation strategy is
// configured.
const gitPullDivergentError = "Need to specify how to reconcile divergent branches"

// gitPullDivergentMatch reports whether git pull failed because the branches
// have diverged and the strategy is not configured.
func gitPullDivergentMatch(command *types.Command) bool {
	return strings.Contains(command.Output, gitPullDivergentError) &&
		utils.Contains(command.ScriptParts(), "pull")
}

// gitPullDivergentNewCommand retries the pull with the strategy that git
// offers in its hint: merge first, then rebase.
func gitPullDivergentNewCommand(command *types.Command) ([]string, error) {
	return []string{
		utils.ReplaceArgument(command.Script, "pull", "pull --no-rebase"),
		utils.ReplaceArgument(command.Script, "pull", "pull --rebase"),
	}, nil
}

func init() {
	register(types.NewRule("git_pull_divergent",
		specific.GitSupportMatch(gitPullDivergentMatch),
		specific.GitSupportNew(gitPullDivergentNewCommand)))
}
