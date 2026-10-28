package rules

import "github.com/leafevans/go-fuck/internal/types"

// cargoMatch reports whether cargo was called without a subcommand.
func cargoMatch(command *types.Command) bool {
	return command.Script == "cargo"
}

// cargoNewCommand builds the project.
func cargoNewCommand(command *types.Command) ([]string, error) {
	return []string{"cargo build"}, nil
}

func init() {
	register(types.NewRule("cargo", cargoMatch, cargoNewCommand))
}
