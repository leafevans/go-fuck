package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/internal/shells"
	"github.com/leafevans/go-fuck/internal/specific"
	"github.com/leafevans/go-fuck/internal/types"
)

// gitPullMatch reports whether git pull failed because the current branch has
// no tracking information.
func gitPullMatch(command *types.Command) bool {
	return strings.Contains(command.Script, "pull") &&
		strings.Contains(command.Output, "set-upstream")
}

// gitPullNewCommand sets the upstream branch suggested by git and retries the
// pull.
func gitPullNewCommand(command *types.Command) ([]string, error) {
	lines := strings.Split(command.Output, "\n")
	if len(lines) < 3 {
		return nil, nil
	}

	// git prints the suggested "git branch --set-upstream-to" command three
	// lines above the end of the output.
	line := strings.TrimSpace(lines[len(lines)-3])
	fields := strings.Split(line, " ")
	branch := fields[len(fields)-1]
	setUpstream := strings.ReplaceAll(line, "<remote>", "origin")
	setUpstream = strings.ReplaceAll(setUpstream, "<branch>", branch)
	return []string{shells.Current().And(setUpstream, command.Script)}, nil
}

func init() {
	register(types.NewRule("git_pull",
		specific.GitSupportMatch(gitPullMatch),
		specific.GitSupportNew(gitPullNewCommand)))
}
