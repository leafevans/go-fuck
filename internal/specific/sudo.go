package specific

import (
	"strings"

	"github.com/leafevans/go-fuck/internal/types"
)

// SudoSupportMatch wraps a match function so that it also matches commands run
// with sudo. The sudo prefix is removed before the wrapped function is called.
func SudoSupportMatch(fn func(*types.Command) bool) func(*types.Command) bool {
	return func(command *types.Command) bool {
		if !strings.HasPrefix(command.Script, "sudo ") {
			return fn(command)
		}
		return fn(command.WithScript(command.Script[5:]))
	}
}

// SudoSupportNew wraps a correction function so that the sudo prefix is
// restored on the corrected commands.
func SudoSupportNew(fn func(*types.Command) ([]string, error)) func(*types.Command) ([]string, error) {
	return func(command *types.Command) ([]string, error) {
		if !strings.HasPrefix(command.Script, "sudo ") {
			return fn(command)
		}
		results, err := fn(command.WithScript(command.Script[5:]))
		if err != nil {
			return results, err
		}
		for i := range results {
			results[i] = "sudo " + results[i]
		}
		return results, nil
	}
}
