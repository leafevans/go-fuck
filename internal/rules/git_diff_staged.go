package rules

import (
	"github.com/leafevans/go-fuck/v3/internal/shells"
	"github.com/leafevans/go-fuck/v3/internal/specific"
	"github.com/leafevans/go-fuck/v3/internal/types"
)

// gitDiffStagedMatch reports whether git diff was called without --staged.
func gitDiffStagedMatch(command *types.Command) bool {
	first := true
	for _, word := range specific.GitCommandWords(command.Script, shells.Current()) {
		if first && word != "diff" {
			return false
		}
		first = false
		switch word {
		case "--staged", "--cached", "--no-index":
			return false
		}
	}
	return !first
}

// gitDiffStagedNewCommand adds the --staged flag.
func gitDiffStagedNewCommand(command *types.Command) ([]string, error) {
	corrected, ok := gitInsertAfterSubcommand(command.Script, "diff", "--staged")
	if !ok {
		return nil, nil
	}
	return []string{corrected}, nil
}

func init() {
	register(types.NewRule("git_diff_staged",
		specific.GitSupportMatch(gitDiffStagedMatch),
		specific.GitSupportNew(gitDiffStagedNewCommand)))
}
