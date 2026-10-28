package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// lsAllMatch reports whether ls printed nothing because the directory only
// contains hidden files.
func lsAllMatch(command *types.Command) bool {
	return strings.TrimSpace(command.Output) == ""
}

// lsAllNewCommand adds the -A flag to show hidden files.
func lsAllNewCommand(command *types.Command) ([]string, error) {
	parts := command.ScriptParts()
	if len(parts) == 0 {
		return nil, nil
	}
	arguments := append([]string{"ls", "-A"}, parts[1:]...)
	return []string{strings.Join(arguments, " ")}, nil
}

func init() {
	register(types.NewRule("ls_all",
		utils.ForApp([]string{"ls"}, 0, lsAllMatch),
		lsAllNewCommand))
}
