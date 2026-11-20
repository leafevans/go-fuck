package rules

import (
	"errors"
	"io/fs"
	"os"
	"strings"

	"github.com/leafevans/go-fuck/v2/internal/shells"
	"github.com/leafevans/go-fuck/v2/internal/types"
)

// Match the English and zh-CN ItemNotFound diagnostics emitted by PowerShell.
var powerShellSetLocationPaths = [...]*lazyRegexp{
	newRegexp(`Cannot find path '([^\r\n]+)' because it does not exist`),
	newRegexp(`\x{627e}\x{4e0d}\x{5230}\x{8def}\x{5f84}\x{201c}([^\r\n]+)\x{201d}\x{ff0c}\x{56e0}\x{4e3a}\x{8be5}\x{8def}\x{5f84}\x{4e0d}\x{5b58}\x{5728}\x{3002}`),
}

// powerShellPathMissing reports whether path does not exist.
var powerShellPathMissing = func(path string) bool {
	_, err := os.Stat(path)
	return errors.Is(err, fs.ErrNotExist)
}

func powerShellMissingLocation(output string) string {
	var path string
	for _, pattern := range powerShellSetLocationPaths {
		if match := pattern.FindStringSubmatch(output); match != nil {
			path = match[1]
			break
		}
	}
	if path == "" || strings.ContainsAny(path, "*?[]") {
		return ""
	}
	// A PowerShell provider or custom drive is not necessarily a directory.
	if colon := strings.IndexByte(path, ':'); colon >= 0 {
		if colon != 1 || len(path) < 3 || !strings.ContainsRune(`\/`, rune(path[2])) ||
			!strings.ContainsRune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ", rune(path[0])) || strings.Contains(path[2:], ":") {
			return ""
		}
	}
	return path
}

// powerShellSetLocationMatch reports whether Set-Location or Push-Location
// failed because the directory is missing.
func powerShellSetLocationMatch(command *types.Command) bool {
	if !powerShellCommandIs(command, "Set-Location", "sl", "chdir", "Push-Location", "pushd") {
		return false
	}
	path := powerShellMissingLocation(command.Output)
	return path != "" && powerShellPathMissing(path)
}

// powerShellSetLocationNewCommand creates the directory and retries.
func powerShellSetLocationNewCommand(command *types.Command) ([]string, error) {
	path := powerShellMissingLocation(command.Output)
	if path == "" {
		return nil, nil
	}
	create := "mkdir -Force " + powerShellQuote(path)
	return []string{shells.Current().And(create, command.Script)}, nil
}

func init() {
	register(types.NewRule("powershell_set_location",
		powerShellSetLocationMatch,
		powerShellSetLocationNewCommand))
}
