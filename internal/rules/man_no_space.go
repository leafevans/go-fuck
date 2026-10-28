package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/internal/types"
)

// manNoSpaceMatch reports whether a man command without a space is used.
func manNoSpaceMatch(command *types.Command) bool {
	return strings.HasPrefix(command.Script, "man") &&
		strings.Contains(strings.ToLower(command.Output), "command not found")
}

// manNoSpaceNewCommand inserts the missing space.
func manNoSpaceNewCommand(command *types.Command) ([]string, error) {
	rest := ""
	if len(command.Script) > len("man") {
		rest = command.Script[len("man"):]
	}
	return []string{"man " + rest}, nil
}

func init() {
	rule := types.NewRule("man_no_space", manNoSpaceMatch, manNoSpaceNewCommand)
	rule.Priority = 2000
	register(rule)
}
