package specific

import (
	"errors"
	"os/exec"
	"slices"
	"strings"

	"github.com/leafevans/go-fuck/v2/internal/shells"
	"github.com/leafevans/go-fuck/v2/internal/utils"
)

// pkgFileResult is the memoized result of a pkgfile lookup.
type pkgFileResult struct {
	packages []string
	err      error
}

// getPkgFile runs pkgfile once per command.
var getPkgFile = utils.Memoize1(lookupPkgFile)

func lookupPkgFile(command string) pkgFileResult {
	if command == "" {
		return pkgFileResult{}
	}
	output, err := exec.Command("pkgfile", "-b", "-v", command).Output()
	return parsePkgFile(output, err)
}

// parsePkgFile keeps process errors separate from a successful empty lookup.
func parsePkgFile(output []byte, err error) pkgFileResult {
	if err != nil {
		var exitError *exec.ExitError
		if errors.As(err, &exitError) && exitError.ExitCode() == 1 && len(output) == 0 && len(exitError.Stderr) == 0 {
			return pkgFileResult{}
		}
		return pkgFileResult{err: err}
	}

	packages := []string{}
	for line := range strings.SplitSeq(strings.TrimSpace(string(output)), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		packages = append(packages, fields[0])
	}
	return pkgFileResult{packages: packages}
}

// GetPkgFile returns the packages that provide the given command, using pkgfile.
//
// When the command has the form "sudo foo" the sudo prefix is ignored.
func GetPkgFile(command string) ([]string, error) {
	parts := shells.Current().SplitCommand(strings.TrimSpace(command))
	if len(parts) > 0 && parts[0] == "sudo" {
		parts = parts[1:]
	}
	if len(parts) == 0 || parts[0] == "" {
		return nil, nil
	}
	result := getPkgFile(parts[0])
	return slices.Clone(result.packages), result.err
}

// ArchlinuxEnv returns the Arch Linux package manager setup.
//
// The first result reports whether the pkgfile based rules are enabled by
// default, the second one is the pacman command to use.
func ArchlinuxEnv() (enabledByDefault bool, pacman string) {
	return archlinuxEnv(utils.IsAvailable)
}

// ArchlinuxDefault reports whether the Arch Linux rules are enabled by default.
func ArchlinuxDefault() bool {
	enabled, _ := archlinuxEnv(utils.StartupAvailable)
	return enabled
}

func archlinuxEnv(available func(string) bool) (enabledByDefault bool, pacman string) {
	switch {
	case available("yay"):
		pacman = "yay"
	case available("pikaur"):
		pacman = "pikaur"
	case available("yaourt"):
		pacman = "yaourt"
	case available("pacman"):
		pacman = "sudo pacman"
	default:
		return false, ""
	}
	return available("pkgfile"), pacman
}
