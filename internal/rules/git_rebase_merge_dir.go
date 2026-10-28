package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/internal/specific"
	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// gitRebaseMergeDirMatch reports whether git rebase failed because another
// rebase is already in progress.
func gitRebaseMergeDirMatch(command *types.Command) bool {
	return strings.Contains(command.Script, " rebase") &&
		strings.Contains(command.Output,
			"It seems that there is already a rebase-merge directory") &&
		strings.Contains(command.Output,
			"I wonder if you are in the middle of another rebase")
}

// gitRebaseMergeDirNewCommand suggests the commands that resolve the pending
// rebase, ordered by similarity to the failed command.
func gitRebaseMergeDirNewCommand(command *types.Command) ([]string, error) {
	commandList := []string{
		"git rebase --continue",
		"git rebase --abort",
		"git rebase --skip",
	}
	lines := strings.Split(command.Output, "\n")
	if len(lines) >= 4 {
		commandList = append(commandList, strings.TrimSpace(lines[len(lines)-4]))
	}
	return utils.GetCloseMatches(command.Script, commandList, 4, 0), nil
}

func init() {
	register(types.NewRule("git_rebase_merge_dir",
		specific.GitSupportMatch(gitRebaseMergeDirMatch),
		specific.GitSupportNew(gitRebaseMergeDirNewCommand)))
}
