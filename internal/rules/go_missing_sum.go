package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/internal/shells"
	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// goMissingSumError matches the Go toolchain error for a module that is
// required but missing from go.sum.
const goMissingSumError = "missing go.sum entry for module providing package"

// goMissingSumMatch reports whether the Go toolchain failed because go.sum is
// incomplete.
func goMissingSumMatch(command *types.Command) bool {
	return strings.Contains(command.Output, goMissingSumError)
}

// goMissingSumNewCommand fills go.sum with go mod tidy and retries the failed
// command.
func goMissingSumNewCommand(command *types.Command) ([]string, error) {
	return []string{shells.Current().And("go mod tidy", command.Script)}, nil
}

func init() {
	register(types.NewRule("go_missing_sum",
		utils.ForApp([]string{"go"}, 0, goMissingSumMatch),
		goMissingSumNewCommand))
}
