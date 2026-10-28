package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// agLiteralMatch reports whether ag suggested searching for a literal string.
func agLiteralMatch(command *types.Command) bool {
	return strings.HasSuffix(command.Output, "run ag with -Q\n")
}

// agLiteralNewCommand adds the -Q flag.
func agLiteralNewCommand(command *types.Command) ([]string, error) {
	return []string{strings.Replace(command.Script, "ag", "ag -Q", 1)}, nil
}

func init() {
	register(types.NewRule("ag_literal",
		utils.ForApp([]string{"ag"}, 0, agLiteralMatch),
		agLiteralNewCommand))
}
