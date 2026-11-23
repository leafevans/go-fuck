package rules

import (
	"github.com/leafevans/go-fuck/v3/internal/shells"
	"github.com/leafevans/go-fuck/v3/internal/specific"
	"github.com/leafevans/go-fuck/v3/internal/types"
)

// gitPushWithoutCommitsRefspec matches the error git prints when the pushed
// refspec has no commits.
var gitPushWithoutCommitsRefspec = newRegexp(`src refspec \w+ does not match any`)

// gitPushWithoutCommitsMatch reports whether git refused to push because
// there is nothing to push yet.
func gitPushWithoutCommitsMatch(command *types.Command) bool {
	return gitPushWithoutCommitsRefspec.MatchString(command.Output)
}

// gitPushWithoutCommitsNewCommand creates the initial commit and retries the
// push.
func gitPushWithoutCommitsNewCommand(command *types.Command) ([]string, error) {
	commit := "git commit -m " + doubleQuote + "Initial commit" + doubleQuote
	return []string{shells.Current().And(commit, command.Script)}, nil
}

func init() {
	register(types.NewRule("git_push_without_commits",
		specific.GitSupportMatch(gitPushWithoutCommitsMatch),
		gitPushWithoutCommitsNewCommand))
}
