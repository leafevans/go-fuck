package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// javaSourceSuffix is the suffix of a Java source file.
const javaSourceSuffix = ".java"

// javaMatch reports whether java was called with a source file instead of a
// class name.
func javaMatch(command *types.Command) bool {
	return strings.HasSuffix(command.Script, javaSourceSuffix)
}

// javaNewCommand drops the file suffix, leaving the class name.
func javaNewCommand(command *types.Command) ([]string, error) {
	// Python: command.script[:-5].
	return []string{strings.TrimSuffix(command.Script, javaSourceSuffix)}, nil
}

func init() {
	register(types.NewRule("java",
		utils.ForApp([]string{"java"}, 0, javaMatch),
		javaNewCommand))
}
