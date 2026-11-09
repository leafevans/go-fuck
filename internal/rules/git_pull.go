package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/internal/shells"
	"github.com/leafevans/go-fuck/internal/specific"
	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// gitPullMatch reports whether git pull failed because the current branch has
// no tracking information.
func gitPullMatch(command *types.Command) bool {
	return strings.Contains(command.Script, "pull") &&
		strings.Contains(command.Output, "set-upstream")
}

// gitPullUpstreamPrefix starts the line with the suggested upstream command.
const gitPullUpstreamPrefix = "git branch --set-upstream-to"

// gitPullNewCommand sets the upstream branch suggested by git and retries the
// pull.
//
// The suggested command is looked up by its prefix instead of its position,
// because the captured output may carry extra empty lines (for example when
// the command was re-run by PowerShell).
func gitPullNewCommand(command *types.Command) ([]string, error) {
	line, found := utils.FindLineByPrefix(command.Output, gitPullUpstreamPrefix)
	if !found {
		return nil, nil
	}
	fields := strings.Fields(line)
	if len(fields) == 0 {
		return nil, nil
	}
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
