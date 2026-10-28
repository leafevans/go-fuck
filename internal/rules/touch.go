package rules

import (
	"regexp"
	"strings"

	"github.com/leafevans/go-fuck/internal/shells"
	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// touchPathPattern extracts the missing directory from the touch error. The
// optional "cannot touch '" prefix is used by GNU coreutils, the BSD
// implementation prints the path directly.
var touchPathPattern = regexp.MustCompile("touch: (cannot touch ')?(.+)/.+?'?:")

// touchMatch reports whether touch failed because the parent directory is
// missing.
func touchMatch(command *types.Command) bool {
	return strings.Contains(command.Output, "No such file or directory")
}

// touchNewCommand creates the missing parent directories.
func touchNewCommand(command *types.Command) ([]string, error) {
	matches := touchPathPattern.FindAllStringSubmatch(command.Output, -1)
	if len(matches) == 0 {
		return nil, nil
	}
	path := matches[0][2]
	return []string{shells.Current().And("mkdir -p "+path, command.Script)}, nil
}

func init() {
	register(types.NewRule("touch",
		utils.ForApp([]string{"touch"}, 0, touchMatch),
		touchNewCommand))
}
