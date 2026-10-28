package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// grepRecursiveMatch reports whether grep failed because it was given a
// directory.
func grepRecursiveMatch(command *types.Command) bool {
	return strings.Contains(strings.ToLower(command.Output), "is a directory")
}

// grepRecursiveNewCommand makes grep search recursively.
func grepRecursiveNewCommand(command *types.Command) ([]string, error) {
	if len(command.Script) < 5 {
		return nil, nil
	}
	return []string{"grep -r " + command.Script[5:]}, nil
}

func init() {
	register(types.NewRule("grep_recursive",
		utils.ForApp([]string{"grep"}, 0, grepRecursiveMatch),
		grepRecursiveNewCommand))
}
