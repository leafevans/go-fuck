package rules

import "github.com/leafevans/go-fuck/internal/types"

// cdParentMatch reports whether cd was called without the space before "..".
func cdParentMatch(command *types.Command) bool {
	return command.Script == "cd.."
}

// cdParentNewCommand adds the missing space.
func cdParentNewCommand(command *types.Command) ([]string, error) {
	return []string{"cd .."}, nil
}

func init() {
	register(types.NewRule("cd_parent", cdParentMatch, cdParentNewCommand))
}
