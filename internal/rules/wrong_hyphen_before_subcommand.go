package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/internal/specific"
	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// wrongHyphenGetAllExecutables returns the installed executables. It is a
// variable so that tests can inject a fixed list.
var wrongHyphenGetAllExecutables = utils.GetAllExecutables

// wrongHyphenMatch reports whether a hyphen was typed instead of a space.
func wrongHyphenMatch(command *types.Command) bool {
	parts := command.ScriptParts()
	if len(parts) == 0 {
		return false
	}

	firstPart := parts[0]
	executables := wrongHyphenGetAllExecutables()
	if !strings.Contains(firstPart, "-") || utils.Contains(executables, firstPart) {
		return false
	}

	cmd := strings.SplitN(firstPart, "-", 2)[0]
	return utils.Contains(executables, cmd)
}

// wrongHyphenNewCommand replaces the first hyphen with a space.
func wrongHyphenNewCommand(command *types.Command) ([]string, error) {
	return []string{strings.Replace(command.Script, "-", " ", 1)}, nil
}

func init() {
	rule := types.NewRule("wrong_hyphen_before_subcommand",
		specific.SudoSupportMatch(wrongHyphenMatch),
		specific.SudoSupportNew(wrongHyphenNewCommand))
	// The rule only inspects the command line.
	rule.RequiresOutput = false
	rule.Priority = 4500
	register(rule)
}
