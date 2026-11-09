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
	// The removal command is looked up by its prefix instead of its position,
	// because the captured output may carry extra empty lines (for example
	// when the command was re-run by PowerShell).
	if line, found := utils.FindLineByPrefix(command.Output, "rm -fr"); found {
		commandList = append(commandList, line)
	}
	return utils.GetCloseMatches(command.Script, commandList, 4, 0), nil
}

func init() {
	register(types.NewRule("git_rebase_merge_dir",
		specific.GitSupportMatch(gitRebaseMergeDirMatch),
		specific.GitSupportNew(gitRebaseMergeDirNewCommand)))
}
