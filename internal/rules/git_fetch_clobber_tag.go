package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/v3/internal/shells"
	"github.com/leafevans/go-fuck/v3/internal/specific"
	"github.com/leafevans/go-fuck/v3/internal/types"
)

// gitFetchClobberTagError matches the rejection that git fetch prints when a
// tag on the remote differs from the local tag.
const gitFetchClobberTagError = "would clobber existing tag"

// gitFetchClobberTagMatch reports whether git fetch refused to update a tag.
func gitFetchClobberTagMatch(command *types.Command) bool {
	return strings.Contains(command.Output, gitFetchClobberTagError) &&
		gitHasSubcommand(command.Script, "fetch")
}

// gitFetchClobberTagNewCommand refetches the tags and overwrites the local
// ones.
func gitFetchClobberTagNewCommand(command *types.Command) ([]string, error) {
	flags := []string{}
	if _, _, ok := specific.GitOptionSpan(command.Script, "--tags", shells.Current()); !ok {
		flags = append(flags, "--tags")
	}
	_, _, longForce := specific.GitOptionSpan(command.Script, "--force", shells.Current())
	_, _, shortForce := specific.GitOptionSpan(command.Script, "-f", shells.Current())
	if !longForce && !shortForce {
		flags = append(flags, "--force")
	}
	if len(flags) == 0 {
		return nil, nil
	}
	corrected, ok := gitInsertAfterSubcommand(command.Script, "fetch", strings.Join(flags, " "))
	if !ok {
		return nil, nil
	}
	return []string{corrected}, nil
}

func init() {
	register(types.NewRule("git_fetch_clobber_tag",
		specific.GitSupportMatch(gitFetchClobberTagMatch),
		specific.GitSupportNew(gitFetchClobberTagNewCommand)))
}
