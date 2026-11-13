package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/v2/internal/specific"
	"github.com/leafevans/go-fuck/v2/internal/types"
	"github.com/leafevans/go-fuck/v2/internal/utils"
)

// gitAmendInitialCommitError matches the error that git prints when --amend is
// used in a repository without commits.
const gitAmendInitialCommitError = "You have nothing to amend"

// gitAmendInitialCommitMatch reports whether git refused to amend because the
// repository has no commits yet.
func gitAmendInitialCommitMatch(command *types.Command) bool {
	return strings.Contains(command.Output, gitAmendInitialCommitError) &&
		utils.Contains(command.ScriptParts(), "--amend")
}

// gitAmendInitialCommitNewCommand drops --amend so that the first commit is
// created normally.
func gitAmendInitialCommitNewCommand(command *types.Command) ([]string, error) {
	fixed := strings.TrimSpace(strings.Replace(command.Script, " --amend", "", 1))
	if fixed == "" || fixed == command.Script {
		return nil, nil
	}
	return []string{fixed}, nil
}

func init() {
	register(types.NewRule("git_amend_initial_commit",
		specific.GitSupportMatch(gitAmendInitialCommitMatch),
		specific.GitSupportNew(gitAmendInitialCommitNewCommand)))
}
