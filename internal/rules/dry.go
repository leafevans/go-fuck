package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/internal/types"
)

// dryMatch reports whether the first two words of the command are equal.
func dryMatch(command *types.Command) bool {
	parts := command.ScriptParts()
	return len(parts) >= 2 && parts[0] == parts[1]
}

// dryNewCommand removes the duplicated word.
func dryNewCommand(command *types.Command) ([]string, error) {
	return []string{strings.Join(command.ScriptParts()[1:], " ")}, nil
}

func init() {
	rule := types.NewRule("dry", dryMatch, dryNewCommand)
	// Typing the same word twice is rare, so the rule comes before rules such
	// as cd_mkdir that also match "cd cd foo".
	rule.Priority = 900
	register(rule)
}
