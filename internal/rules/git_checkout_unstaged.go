package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/internal/shells"
	"github.com/leafevans/go-fuck/internal/specific"
	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// gitCheckoutUnstagedError matches the error that git prints when a checkout
// or switch would overwrite local changes.
const gitCheckoutUnstagedError = "would be overwritten by checkout"

// gitCheckoutUnstagedMatch reports whether git checkout or git switch failed
// because of local changes.
func gitCheckoutUnstagedMatch(command *types.Command) bool {
	parts := command.ScriptParts()
	if !utils.Contains(parts, "checkout") && !utils.Contains(parts, "switch") {
		return false
	}
	return strings.Contains(command.Output, gitCheckoutUnstagedError) &&
		strings.Contains(command.Output, "or stash them")
}

// gitCheckoutUnstagedNewCommand stashes the changes, switches and restores
// them.
func gitCheckoutUnstagedNewCommand(command *types.Command) ([]string, error) {
	return []string{shells.Current().And("git stash", command.Script, "git stash pop")}, nil
}

func init() {
	rule := types.NewRule("git_checkout_unstaged",
		specific.GitSupportMatch(gitCheckoutUnstagedMatch),
		specific.GitSupportNew(gitCheckoutUnstagedNewCommand))
	// Deviation: git_stash of the reference implementation matches the same
	// output and suggests a stash without restoring the changes afterwards.
	// This rule ranks first so that the complete correction is suggested above
	// the plain one.
	rule.Priority = 900
	register(rule)
}
