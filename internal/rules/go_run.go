package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// goRunSuffix is the suffix that go run expects for source files.
const goRunSuffix = ".go"

// goRunMatch reports whether "go run" was called without a .go suffix.
func goRunMatch(command *types.Command) bool {
	return strings.HasPrefix(command.Script, "go run ") &&
		!strings.HasSuffix(command.Script, goRunSuffix)
}

// goRunNewCommand appends the missing suffix.
func goRunNewCommand(command *types.Command) ([]string, error) {
	return []string{command.Script + goRunSuffix}, nil
}

func init() {
	register(types.NewRule("go_run",
		utils.ForApp([]string{"go"}, 0, goRunMatch),
		goRunNewCommand))
}
