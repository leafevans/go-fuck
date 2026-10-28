package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// lsLahMatch reports whether ls was called without the -lah flags.
func lsLahMatch(command *types.Command) bool {
	parts := command.ScriptParts()
	return len(parts) > 0 && !strings.Contains(command.Script, "ls -")
}

// lsLahNewCommand adds the -lah flags.
func lsLahNewCommand(command *types.Command) ([]string, error) {
	parts := command.ScriptParts()
	if len(parts) == 0 {
		return nil, nil
	}
	arguments := append([]string{}, parts...)
	arguments[0] = "ls -lah"
	return []string{strings.Join(arguments, " ")}, nil
}

func init() {
	register(types.NewRule("ls_lah",
		utils.ForApp([]string{"ls"}, 0, lsLahMatch),
		lsLahNewCommand))
}
