package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/v4/internal/shells"
	"github.com/leafevans/go-fuck/v4/internal/specific"
	"github.com/leafevans/go-fuck/v4/internal/types"
)

// Git versions differ on whether the question mark is inside the parentheses.
var gitTwoDashesHint = newRegexp("(?m)^error: did you mean `(--[A-Za-z0-9][A-Za-z0-9-]*)` \\(with two dashes(?: \\?\\)|\\)\\?)\\r?$")

func gitTwoDashesCorrection(command *types.Command) (start, end int, option string, ok bool) {
	if !strings.Contains(command.Output, "with two dashes") {
		return 0, 0, "", false
	}
	hint := gitTwoDashesHint.FindStringSubmatch(command.Output)
	if hint == nil {
		return 0, 0, "", false
	}
	option = hint[1]
	start, end, ok = specific.GitOptionSpan(command.Script, option[1:], shells.Current())
	return start, end, option, ok
}

// gitTwoDashesMatch reports whether git suggested a long option instead of a
// single dash one.
func gitTwoDashesMatch(command *types.Command) bool {
	_, _, _, ok := gitTwoDashesCorrection(command)
	return ok
}

// gitTwoDashesNewCommand replaces the single dash option with the long option
// suggested by git.
func gitTwoDashesNewCommand(command *types.Command) ([]string, error) {
	start, end, option, ok := gitTwoDashesCorrection(command)
	if !ok {
		return nil, nil
	}
	return []string{command.Script[:start] + option + command.Script[end:]}, nil
}

func init() {
	register(types.NewRule("git_two_dashes",
		specific.GitSupportMatch(gitTwoDashesMatch),
		specific.GitSupportNew(gitTwoDashesNewCommand)))
}
