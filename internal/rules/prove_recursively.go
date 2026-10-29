package rules

import (
	"os"
	"slices"
	"strings"

	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// proveRecursivelyIsDir reports whether the argument is a directory. It is a
// variable so that tests do not depend on the file system.
var proveRecursivelyIsDir = func(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// proveRecursivelyIsRecursive reports whether the argument enables recursive
// mode.
func proveRecursivelyIsRecursive(part string) bool {
	if part == "--recurse" {
		return true
	}
	return !strings.HasPrefix(part, "--") &&
		strings.HasPrefix(part, "-") &&
		strings.Contains(part, "r")
}

// proveRecursivelyMatch reports whether prove was run without -r on a
// directory.
func proveRecursivelyMatch(command *types.Command) bool {
	if !strings.Contains(command.Output, "NOTESTS") {
		return false
	}

	parts := command.ScriptParts()[1:]
	if slices.ContainsFunc(parts, proveRecursivelyIsRecursive) {
		return false
	}
	for _, part := range parts {
		if !strings.HasPrefix(part, "-") && proveRecursivelyIsDir(part) {
			return true
		}
	}
	return false
}

// proveRecursivelyNewCommand enables recursive mode.
func proveRecursivelyNewCommand(command *types.Command) ([]string, error) {
	parts := command.ScriptParts()
	if len(parts) == 0 {
		return nil, nil
	}
	result := make([]string, 0, len(parts)+1)
	result = append(result, parts[0], "-r")
	result = append(result, parts[1:]...)
	return []string{strings.Join(result, " ")}, nil
}

func init() {
	register(types.NewRule("prove_recursively",
		utils.ForApp([]string{"prove"}, 0, proveRecursivelyMatch),
		proveRecursivelyNewCommand))
}
