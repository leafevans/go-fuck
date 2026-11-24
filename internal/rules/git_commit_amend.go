package rules

import (
	"github.com/leafevans/go-fuck/v3/internal/specific"
	"github.com/leafevans/go-fuck/v3/internal/types"
)

// gitCommitAmendMatch reports whether a commit command was run and rewriting the
// previous commit can fix the failure.
func gitCommitAmendMatch(command *types.Command) bool {
	return gitHasSubcommand(command.Script, "commit") && gitCommitRewritable(command)
}

// gitCommitAmendNewCommand amends the previous commit.
func gitCommitAmendNewCommand(command *types.Command) ([]string, error) {
	return []string{"git commit --amend"}, nil
}

func init() {
	rule := types.NewRule("git_commit_amend",
		specific.GitSupportMatch(gitCommitAmendMatch),
		specific.GitSupportNew(gitCommitAmendNewCommand))
	// Deviation: the reference implementation gives this rule the default
	// priority, so it is suggested before any concrete fix. Amending is only a
	// guess when a commit fails; concrete rules such as git_add and
	// git_hook_bypass must be offered first.
	// No generic commit failure proves that rewriting history is the fix.
	// Keep this legacy convenience rule available only by explicit opt-in.
	rule.EnabledByDefault = false
	rule.Reason = "Rewrite the previous commit; this changes local Git history"
	rule.Priority = 1500
	register(rule)
}
