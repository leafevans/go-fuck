package specific

import (
	"errors"
	"os/exec"
	"strings"

	"github.com/leafevans/go-fuck/internal/utils"
)

// pkgFileResult is the memoized result of a pkgfile lookup.
type pkgFileResult struct {
	packages []string
	err      error
}

// getPkgFile runs pkgfile once per command.
var getPkgFile = utils.Memoize1(func(command string) pkgFileResult {
	command = strings.TrimSpace(command)
	command = strings.TrimPrefix(command, "sudo ")
	fields := strings.Fields(command)
	if len(fields) == 0 {
		return pkgFileResult{}
	}

	output, err := exec.Command("pkgfile", "-b", "-v", fields[0]).Output()
	if err != nil {
		var exitError *exec.ExitError
		if errors.As(err, &exitError) && exitError.ExitCode() == 1 && len(output) == 0 {
			return pkgFileResult{}
		}
		return pkgFileResult{err: err}
	}

	packages := []string{}
	for _, line := range strings.Split(strings.TrimSpace(string(output)), "\n") {
		if line == "" {
			continue
		}
		packages = append(packages, strings.Fields(line)[0])
	}
	return pkgFileResult{packages: packages}
})

// GetPkgFile returns the packages that provide the given command, using pkgfile.
//
// When the command has the form "sudo foo" the sudo prefix is ignored.
func GetPkgFile(command string) ([]string, error) {
	result := getPkgFile(command)
	return result.packages, result.err
}

// ArchlinuxEnv returns the Arch Linux package manager setup.
//
// The first result reports whether the pkgfile based rules are enabled by
// default, the second one is the pacman command to use.
func ArchlinuxEnv() (enabledByDefault bool, pacman string) {
	switch {
	case utils.Which("yay") != "":
		pacman = "yay"
	case utils.Which("pikaur") != "":
		pacman = "pikaur"
	case utils.Which("yaourt") != "":
		pacman = "yaourt"
	case utils.Which("pacman") != "":
		pacman = "sudo pacman"
	default:
		return false, ""
	}
	return utils.Which("pkgfile") != "", pacman
}
