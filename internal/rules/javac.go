package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// javacSuffix is the suffix that javac expects for source files.
const javacSuffix = ".java"

// javacMatch reports whether javac was called without a .java suffix.
func javacMatch(command *types.Command) bool {
	return !strings.HasSuffix(command.Script, javacSuffix)
}

// javacNewCommand appends the missing suffix.
func javacNewCommand(command *types.Command) ([]string, error) {
	return []string{command.Script + javacSuffix}, nil
}

func init() {
	register(types.NewRule("javac",
		utils.ForApp([]string{"javac"}, 0, javacMatch),
		javacNewCommand))
}
