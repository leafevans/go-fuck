package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/internal/specific"
	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// gitRmRecursiveMatch reports whether git rm refused to delete a directory
// without the recursive flag.
func gitRmRecursiveMatch(command *types.Command) bool {
	return strings.Contains(command.Script, " rm ") &&
		strings.Contains(command.Output, "fatal: not removing '") &&
		strings.Contains(command.Output, "' recursively without -r")
}

// gitRmRecursiveNewCommand adds the -r flag.
func gitRmRecursiveNewCommand(command *types.Command) ([]string, error) {
	parts := append([]string{}, command.ScriptParts()...)
	index := utils.Index(parts, "rm")
	if index < 0 {
		return nil, nil
	}
	index++

	newParts := make([]string, 0, len(parts)+1)
	newParts = append(newParts, parts[:index]...)
	newParts = append(newParts, "-r")
	newParts = append(newParts, parts[index:]...)
	return []string{strings.Join(newParts, " ")}, nil
}

func init() {
	register(types.NewRule("git_rm_recursive",
		specific.GitSupportMatch(gitRmRecursiveMatch),
		specific.GitSupportNew(gitRmRecursiveNewCommand)))
}
