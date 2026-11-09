package rules

import (
	"os"
	"regexp"

	"github.com/leafevans/go-fuck/internal/shells"
	"github.com/leafevans/go-fuck/internal/types"
)

// powerShellSetLocationPath matches the path of a failed Set-Location.
var powerShellSetLocationPath = regexp.MustCompile(`Cannot find path '([^']*)' because it does not exist`)

// powerShellPathMissing reports whether path does not exist.
var powerShellPathMissing = func(path string) bool {
	_, err := os.Stat(path)
	return err != nil
}

// powerShellSetLocationMatch reports whether Set-Location failed because the
// directory is missing.
func powerShellSetLocationMatch(command *types.Command) bool {
	if !powerShellCommandIs(command, "Set-Location", "sl", "chdir") {
		return false
	}
	match := powerShellSetLocationPath.FindStringSubmatch(command.Output)
	if match == nil {
		return false
	}
	return powerShellPathMissing(match[1])
}

// powerShellSetLocationNewCommand creates the directory and retries.
func powerShellSetLocationNewCommand(command *types.Command) ([]string, error) {
	match := powerShellSetLocationPath.FindStringSubmatch(command.Output)
	if match == nil || match[1] == "" {
		return nil, nil
	}
	create := "mkdir -Force " + powerShellQuote(match[1])
	return []string{shells.Current().And(create, command.Script)}, nil
}

func init() {
	register(types.NewRule("powershell_set_location",
		powerShellSetLocationMatch,
		powerShellSetLocationNewCommand))
}
