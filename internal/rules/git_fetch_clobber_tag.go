package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/internal/specific"
	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// gitFetchClobberTagError matches the rejection that git fetch prints when a
// tag on the remote differs from the local tag.
const gitFetchClobberTagError = "would clobber existing tag"

// gitFetchClobberTagMatch reports whether git fetch refused to update a tag.
func gitFetchClobberTagMatch(command *types.Command) bool {
	return strings.Contains(command.Output, gitFetchClobberTagError) &&
		utils.Contains(command.ScriptParts(), "fetch")
}

// gitFetchClobberTagNewCommand refetches the tags and overwrites the local
// ones.
func gitFetchClobberTagNewCommand(command *types.Command) ([]string, error) {
	parts := command.ScriptParts()
	arguments := "fetch"
	if !utils.Contains(parts, "--tags") {
		arguments += " --tags"
	}
	if !utils.Contains(parts, "--force") && !utils.Contains(parts, "-f") {
		arguments += " --force"
	}
	if arguments == "fetch" {
		return nil, nil
	}
	return []string{utils.ReplaceArgument(command.Script, "fetch", arguments)}, nil
}

func init() {
	register(types.NewRule("git_fetch_clobber_tag",
		specific.GitSupportMatch(gitFetchClobberTagMatch),
		specific.GitSupportNew(gitFetchClobberTagNewCommand)))
}
