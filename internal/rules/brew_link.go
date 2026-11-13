package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// brewLinkMatch reports whether brew could not link a formula because the
// target files already exist.
func brewLinkMatch(command *types.Command) bool {
	parts := command.ScriptParts()
	if len(parts) < 2 {
		return false
	}
	action := parts[1]
	return (action == "ln" || action == "link") &&
		strings.Contains(command.Output, "brew link --overwrite --dry-run")
}

// brewLinkNewCommand links the formula while listing the files that would be
// deleted.
func brewLinkNewCommand(command *types.Command) ([]string, error) {
	parts := command.ScriptParts()
	if len(parts) < 2 {
		return nil, nil
	}
	newParts := append([]string{}, parts[:2]...)
	newParts = append(newParts, "--overwrite", "--dry-run")
	newParts = append(newParts, parts[2:]...)
	return []string{strings.Join(newParts, " ")}, nil
}

func init() {
	register(types.NewRule("brew_link",
		utils.ForApp([]string{"brew"}, 2, brewLinkMatch),
		brewLinkNewCommand))
}
