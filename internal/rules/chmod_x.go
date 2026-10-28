package rules

import (
	"os"
	"runtime"
	"strings"

	"github.com/leafevans/go-fuck/internal/shells"
	"github.com/leafevans/go-fuck/internal/types"
)

// chmodXPathExists reports whether a path exists. It is a variable so that
// tests can replace the file system lookup.
var chmodXPathExists = func(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// chmodXIsExecutable reports whether a path is executable. It is a variable so
// that tests can replace the file system lookup.
var chmodXIsExecutable = func(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	// Python's os.access(path, os.X_OK) treats every existing file as
	// executable on Windows, where there are no execute permission bits.
	if runtime.GOOS == "windows" {
		return true
	}
	return info.Mode()&0o111 != 0
}

// chmodXMatch reports whether a script failed because it is not executable.
func chmodXMatch(command *types.Command) bool {
	parts := command.ScriptParts()
	if len(parts) == 0 || !strings.HasPrefix(command.Script, "./") {
		return false
	}
	if !strings.Contains(strings.ToLower(command.Output), "permission denied") {
		return false
	}
	return chmodXPathExists(parts[0]) && !chmodXIsExecutable(parts[0])
}

// chmodXNewCommand makes the script executable before running it again.
func chmodXNewCommand(command *types.Command) ([]string, error) {
	parts := command.ScriptParts()
	if len(parts) == 0 || len(parts[0]) < 2 {
		return nil, nil
	}
	script := shells.Current().And("chmod +x "+parts[0][2:], command.Script)
	return []string{script}, nil
}

func init() {
	register(types.NewRule("chmod_x", chmodXMatch, chmodXNewCommand))
}
