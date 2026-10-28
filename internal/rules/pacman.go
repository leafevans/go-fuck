package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/internal/shells"
	"github.com/leafevans/go-fuck/internal/specific"
	"github.com/leafevans/go-fuck/internal/types"
)

// pacmanPkgFile returns the packages that provide a command, using pkgfile.
var pacmanPkgFile = specific.GetPkgFile

// pacmanInstallCommand returns the command used to install packages. It is a
// variable so that the tests do not depend on the host.
var pacmanInstallCommand = func() string {
	_, pacman := specific.ArchlinuxEnv()
	return pacman
}

// pacmanPackages returns the packages that provide the command.
func pacmanPackages(command *types.Command) []string {
	packages, err := pacmanPkgFile(command.Script)
	if err != nil {
		return nil
	}
	return packages
}

// pacmanMatch reports whether a missing executable is provided by an Arch
// Linux package.
func pacmanMatch(command *types.Command) bool {
	return strings.Contains(command.Output, "not found") && len(pacmanPackages(command)) > 0
}

// pacmanNewCommand installs the missing packages and re-runs the command.
func pacmanNewCommand(command *types.Command) ([]string, error) {
	install := pacmanInstallCommand()
	results := []string{}
	for _, packageName := range pacmanPackages(command) {
		results = append(results, shells.Current().And(install+" -S "+packageName, command.Script))
	}
	return results, nil
}

func init() {
	rule := types.NewRule("pacman", pacmanMatch, pacmanNewCommand)
	rule.EnabledByDefault, _ = specific.ArchlinuxEnv()
	register(rule)
}
