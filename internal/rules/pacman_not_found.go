package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/internal/specific"
	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// pacmanNotFoundPkgFile returns the packages that provide a command, using
// pkgfile.
var pacmanNotFoundPkgFile = specific.GetPkgFile

// pacmanNotFoundApps lists the Arch Linux package managers supported by the
// rule.
var pacmanNotFoundApps = []string{"pacman", "yay", "pikaur", "yaourt"}

// pacmanNotFoundMatch reports whether a package manager could not find a
// package.
func pacmanNotFoundMatch(command *types.Command) bool {
	parts := command.ScriptParts()
	if len(parts) == 0 || !strings.Contains(command.Output, "error: target not found:") {
		return false
	}
	if utils.Contains(pacmanNotFoundApps, parts[0]) {
		return true
	}
	return len(parts) >= 2 && parts[0] == "sudo" && parts[1] == "pacman"
}

// pacmanNotFoundNewCommand suggests the packages that provide the executable
// with the same name.
func pacmanNotFoundNewCommand(command *types.Command) ([]string, error) {
	parts := command.ScriptParts()
	if len(parts) == 0 {
		return nil, nil
	}
	broken := parts[len(parts)-1]
	packages, err := pacmanNotFoundPkgFile(broken)
	if err != nil {
		return nil, nil
	}
	return utils.ReplaceCommand(command, broken, packages), nil
}

func init() {
	rule := types.NewRule("pacman_not_found", pacmanNotFoundMatch, pacmanNotFoundNewCommand)
	rule.EnabledByDefault, _ = specific.ArchlinuxEnv()
	register(rule)
}
