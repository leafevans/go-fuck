package rules

import "github.com/leafevans/go-fuck/internal/types"

// slLsMatch reports whether sl was typed instead of ls.
func slLsMatch(command *types.Command) bool {
	return command.Script == "sl"
}

// slLsNewCommand suggests ls.
func slLsNewCommand(command *types.Command) ([]string, error) {
	return []string{"ls"}, nil
}

func init() {
	register(types.NewRule("sl_ls", slLsMatch, slLsNewCommand))
}
