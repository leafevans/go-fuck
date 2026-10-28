package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/internal/specific"
	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// gitDiffNoIndexMatch reports whether two files were compared without the
// --no-index flag.
func gitDiffNoIndexMatch(command *types.Command) bool {
	parts := command.ScriptParts()
	files := 0
	for index := 2; index < len(parts); index++ {
		if !strings.HasPrefix(parts[index], "-") {
			files++
		}
	}
	return strings.Contains(command.Script, "diff") &&
		!strings.Contains(command.Script, "--no-index") &&
		files == 2
}

// gitDiffNoIndexNewCommand adds the --no-index flag.
func gitDiffNoIndexNewCommand(command *types.Command) ([]string, error) {
	return []string{utils.ReplaceArgument(command.Script, "diff", "diff --no-index")}, nil
}

func init() {
	register(types.NewRule("git_diff_no_index",
		specific.GitSupportMatch(gitDiffNoIndexMatch),
		specific.GitSupportNew(gitDiffNoIndexNewCommand)))
}
