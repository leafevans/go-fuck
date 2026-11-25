package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/v4/internal/specific"
	"github.com/leafevans/go-fuck/v4/internal/types"
)

// gitCommitRewriteBlockers list the commit failures that rewriting the previous
// commit does not fix: a hook rejected the commit, git cannot identify the
// author, a merge is unresolved, or nothing was staged, so no new commit exists
// to replace.
var gitCommitRewriteBlockers = []string{
	"author identity unknown",
	"unable to auto-detect email address",
	"committer identity unknown",
	"unmerged files",
	"you have not concluded your merge",
	"fix conflicts",
	"nothing to commit",
	"no changes added to commit",
	"nothing added to commit",
}

// gitCommitRewritable reports whether rewriting the previous commit can fix a
// failed commit.
//
// Deviation: the reference implementation offers the history rewrite for every
// failed commit, including failures with a concrete cause such as a rejected
// hook, a missing commit identity or unresolved merge conflicts.
func gitCommitRewritable(command *types.Command) bool {
	if gitHookFailure(command.Output) {
		return false
	}
	output := strings.ToLower(command.Output)
	for _, blocker := range gitCommitRewriteBlockers {
		if strings.Contains(output, blocker) {
			return false
		}
	}
	return true
}

// gitCommitResetMatch reports whether a commit command was run and rewriting the
// previous commit can fix the failure.
func gitCommitResetMatch(command *types.Command) bool {
	return gitHasSubcommand(command.Script, "commit") && gitCommitRewritable(command)
}

// gitCommitResetNewCommand resets the last commit.
func gitCommitResetNewCommand(command *types.Command) ([]string, error) {
	return []string{"git reset HEAD~"}, nil
}

func init() {
	rule := types.NewRule("git_commit_reset",
		specific.GitSupportMatch(gitCommitResetMatch),
		specific.GitSupportNew(gitCommitResetNewCommand))
	// Deviation: the reference implementation gives this rule the default
	// priority, so it is suggested before any concrete fix. Resetting the last
	// commit is only a guess when a commit fails and is destructive; concrete
	// rules must be offered first.
	// No generic commit failure proves that rewriting history is the fix.
	// Keep this legacy convenience rule available only by explicit opt-in.
	rule.EnabledByDefault = false
	rule.Reason = "Rewrite the previous commit; this changes local Git history"
	rule.Priority = 1500
	register(rule)
}
