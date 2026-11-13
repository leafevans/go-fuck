package rules

import (
	"regexp"
	"strings"

	"github.com/leafevans/go-fuck/v2/internal/types"
)

// cmdExistingDirectory matches the existing item that mkdir reports.
var cmdExistingDirectory = regexp.MustCompile(`A subdirectory or file (.+) already exists\.`)

// cmdMkdirExistingMatch reports whether mkdir failed because a directory with
// the same name already exists.
func cmdMkdirExistingMatch(command *types.Command) bool {
	if !strings.EqualFold(cmdFirstWord(command.Script), "mkdir") &&
		!strings.EqualFold(cmdFirstWord(command.Script), "md") {
		return false
	}
	if len(cmdSplit(command.Script)) != 2 {
		return false
	}
	match := cmdExistingDirectory.FindStringSubmatch(command.Output)
	if match == nil {
		return false
	}
	return cmdPathIsDirectory(match[1])
}

// cmdMkdirExistingNewCommand enters the existing directory.
func cmdMkdirExistingNewCommand(command *types.Command) ([]string, error) {
	match := cmdExistingDirectory.FindStringSubmatch(command.Output)
	if match == nil {
		return nil, nil
	}
	return []string{"cd " + cmdQuote(match[1])}, nil
}

func init() {
	register(types.NewRule("cmd_mkdir_existing",
		cmdMkdirExistingMatch,
		cmdMkdirExistingNewCommand))
}
