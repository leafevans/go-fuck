package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/internal/specific"
	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// gitTwoDashesMatch reports whether git suggested a long option instead of a
// single dash one.
func gitTwoDashesMatch(command *types.Command) bool {
	return strings.Contains(command.Output, "error: did you mean `") &&
		strings.Contains(command.Output, "` (with two dashes ?)")
}

// gitTwoDashesNewCommand replaces the single dash option with the long option
// suggested by git.
func gitTwoDashesNewCommand(command *types.Command) ([]string, error) {
	parts := strings.Split(command.Output, "`")
	if len(parts) < 2 || len(parts[1]) < 2 {
		return nil, nil
	}
	to := parts[1]
	return []string{utils.ReplaceArgument(command.Script, to[1:], to)}, nil
}

func init() {
	register(types.NewRule("git_two_dashes",
		specific.GitSupportMatch(gitTwoDashesMatch),
		specific.GitSupportNew(gitTwoDashesNewCommand)))
}
