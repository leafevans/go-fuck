package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/internal/types"
)

// quotationMarksMatch reports whether a command mixes single and double quotes.
func quotationMarksMatch(command *types.Command) bool {
	return strings.ContainsRune(command.Script, '\'') &&
		strings.ContainsRune(command.Script, '"')
}

// quotationMarksNewCommand replaces single quotes with double quotes.
func quotationMarksNewCommand(command *types.Command) ([]string, error) {
	return []string{strings.ReplaceAll(command.Script, "'", doubleQuote)}, nil
}

func init() {
	register(types.NewRule("quotation_marks",
		quotationMarksMatch,
		quotationMarksNewCommand))
}
