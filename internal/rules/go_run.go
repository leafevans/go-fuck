package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// goRunSuffix is the suffix that go run expects for source files.
const goRunSuffix = ".go"

// goRunMatch reports whether "go run" was called with a single file argument
// that has no extension at all. A mistyped extension is corrected by
// go_missing_file, so appending ".go" to it would only produce a wrong name.
func goRunMatch(command *types.Command) bool {
	if !strings.HasPrefix(command.Script, "go run ") ||
		strings.HasSuffix(command.Script, goRunSuffix) {
		return false
	}
	argument := command.Script[len("go run "):]
	return argument != "" &&
		!strings.Contains(argument, ".") &&
		!strings.ContainsAny(argument, " \t")
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
