package rules

import (
	"os"
	"strings"

	"github.com/leafevans/go-fuck/v3/internal/types"
	"github.com/leafevans/go-fuck/v3/internal/utils"
)

// grepArgumentsOrderPathExists reports whether a command argument names a file
// or a directory. It is a variable so that tests can replace the file system
// lookup.
var grepArgumentsOrderPathExists = func(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// grepArgumentsOrderActualFile returns the first argument that is a file or a
// directory.
func grepArgumentsOrderActualFile(parts []string) string {
	if len(parts) == 0 {
		return ""
	}
	for _, part := range parts[1:] {
		if grepArgumentsOrderPathExists(part) {
			return part
		}
	}
	return ""
}

// grepArgumentsOrderMatch reports whether grep failed because the searched
// pattern and the file argument are swapped.
func grepArgumentsOrderMatch(command *types.Command) bool {
	return strings.Contains(command.Output, ": No such file or directory") &&
		grepArgumentsOrderActualFile(command.ScriptParts()) != ""
}

// grepArgumentsOrderNewCommand moves the file to the end of the command.
func grepArgumentsOrderNewCommand(command *types.Command) ([]string, error) {
	actualFile := grepArgumentsOrderActualFile(command.ScriptParts())
	if actualFile == "" {
		return nil, nil
	}

	parts := append([]string{}, command.ScriptParts()...)
	if index := utils.Index(parts, actualFile); index >= 0 {
		parts = append(parts[:index], parts[index+1:]...)
	}
	parts = append(parts, actualFile)
	suggestion := strings.Join(parts, " ")
	if suggestion == strings.Join(command.ScriptParts(), " ") {
		// The file argument is already last, so the command cannot be
		// improved and the rule must not offer the same command again.
		return nil, nil
	}
	return []string{suggestion}, nil
}

func init() {
	register(types.NewRule("grep_arguments_order",
		utils.ForApp([]string{"grep", "egrep"}, 0, grepArgumentsOrderMatch),
		grepArgumentsOrderNewCommand))
}
