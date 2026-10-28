package rules

import (
	"os"
	"strings"

	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// catDirIsDir reports whether the path is a directory. It is a variable so that
// tests can replace the file system lookup.
var catDirIsDir = func(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// catDirMatch reports whether cat was called on a directory.
func catDirMatch(command *types.Command) bool {
	parts := command.ScriptParts()
	return len(parts) > 1 &&
		strings.HasPrefix(command.Output, "cat: ") &&
		catDirIsDir(parts[1])
}

// catDirNewCommand calls ls on the directory instead.
func catDirNewCommand(command *types.Command) ([]string, error) {
	return []string{strings.Replace(command.Script, "cat", "ls", 1)}, nil
}

func init() {
	register(types.NewRule("cat_dir",
		utils.ForApp([]string{"cat"}, 1, catDirMatch),
		catDirNewCommand))
}
