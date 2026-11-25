package rules

import (
	"strings"
	"unicode/utf8"

	"github.com/leafevans/go-fuck/v4/internal/types"
	"github.com/leafevans/go-fuck/v4/internal/utils"
)

// missingSpaceGetAllExecutables returns the installed executables. It is a
// variable so that tests can inject a fixed list.
var missingSpaceGetAllExecutables = utils.GetAllExecutables

// missingSpaceExecutable returns the installed executable that the script part
// starts with, or an empty string when there is none.
func missingSpaceExecutable(scriptPart string) string {
	for _, executable := range missingSpaceGetAllExecutables() {
		// The prefix check fails immediately for almost every executable, so it
		// runs before the rune count.
		if strings.HasPrefix(scriptPart, executable) && utf8.RuneCountInString(executable) > 1 {
			return executable
		}
	}
	return ""
}

// missingSpaceMatch reports whether a subcommand is missing its space.
func missingSpaceMatch(command *types.Command) bool {
	parts := command.ScriptParts()
	if len(parts) == 0 {
		return false
	}
	if utils.Contains(missingSpaceGetAllExecutables(), parts[0]) {
		return false
	}
	return missingSpaceExecutable(parts[0]) != ""
}

// missingSpaceNewCommand inserts the missing space.
func missingSpaceNewCommand(command *types.Command) ([]string, error) {
	parts := command.ScriptParts()
	if len(parts) == 0 {
		return nil, nil
	}
	executable := missingSpaceExecutable(parts[0])
	if executable == "" {
		return nil, nil
	}
	return []string{strings.Replace(command.Script, executable, executable+" ", 1)}, nil
}

func init() {
	rule := types.NewRule("missing_space_before_subcommand", missingSpaceMatch, missingSpaceNewCommand)
	rule.Priority = 4000
	register(rule)
}
