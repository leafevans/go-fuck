package rules

import (
	"os"
	"os/exec"
	"strings"

	"github.com/leafevans/go-fuck/internal/specific"
	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// aptGetHandlerPath returns the path of the command-not-found helper script
// that maps an executable to the packages providing it, or an empty string
// when the helper is not available.
//
// Deviation: the original rule uses the python-apt based CommandNotFound
// module to read the command database. go-fuck has no access to python-apt, so
// it queries the same database through the command-not-found helper shipped by
// Debian and Ubuntu, which is the command line equivalent of that module.
var aptGetHandlerPath = func() string {
	if path := utils.Which("command-not-found"); path != "" {
		return path
	}
	const handler = "/usr/lib/command-not-found"
	if _, err := os.Stat(handler); err == nil {
		return handler
	}
	return ""
}

// aptGetWhich returns the path of an executable. It is a variable so that the
// tests do not depend on the host.
var aptGetWhich = utils.Which

// aptGetPackages returns the names of the packages that provide the
// executable, in the order reported by command-not-found.
var aptGetPackages = utils.Memoize1(func(executable string) []string {
	handler := aptGetHandlerPath()
	if handler == "" {
		return nil
	}
	output, _ := exec.Command(handler, executable).CombinedOutput()
	packages := []string{}
	for _, line := range strings.Split(string(output), "\n") {
		line = strings.TrimSpace(line)
		for _, prefix := range []string{"sudo apt install ", "sudo apt-get install "} {
			if strings.HasPrefix(line, prefix) {
				packages = append(packages, strings.TrimSpace(strings.TrimPrefix(line, prefix)))
			}
		}
	}
	return packages
})

// aptGetExecutable returns the executable that the command tried to run.
func aptGetExecutable(command *types.Command) string {
	parts := command.ScriptParts()
	if len(parts) == 0 {
		return ""
	}
	if parts[0] == "sudo" {
		if len(parts) < 2 {
			return ""
		}
		return parts[1]
	}
	return parts[0]
}

// aptGetPackage returns the package that provides the executable, or an empty
// string when no package was found.
func aptGetPackage(executable string) string {
	packages := aptGetPackages(executable)
	if len(packages) == 0 {
		return ""
	}
	return packages[0]
}

// aptGetMatch reports whether a missing executable can be installed with apt.
func aptGetMatch(command *types.Command) bool {
	if !strings.Contains(command.Output, "not found") &&
		!strings.Contains(command.Output, "not installed") {
		return false
	}
	executable := aptGetExecutable(command)
	if executable == "" || aptGetWhich(executable) != "" {
		return false
	}
	return aptGetPackage(executable) != ""
}

// aptGetNewCommand installs the missing package and re-runs the command.
func aptGetNewCommand(command *types.Command) ([]string, error) {
	packageName := aptGetPackage(aptGetExecutable(command))
	if packageName == "" {
		return nil, nil
	}
	install := "sudo apt-get install " + packageName
	return []string{install + " && " + command.Script}, nil
}

func init() {
	rule := types.NewRule("apt_get", aptGetMatch, aptGetNewCommand)
	rule.EnabledByDefault = specific.AptAvailable() && aptGetHandlerPath() != ""
	register(rule)
}
