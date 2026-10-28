package rules

import (
	"regexp"
	"strings"

	"github.com/leafevans/go-fuck/internal/shells"
	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// nodeMissingModule matches the module name in the Node.js error.
var nodeMissingModule = regexp.MustCompile(`Cannot find module '([^']*)'`)

// nodeMissingModuleMatch matches the error that Node.js prints when a module
// cannot be resolved.
func nodeMissingModuleMatch(command *types.Command) bool {
	output := command.Output
	return strings.Contains(output, "Cannot find module '") &&
		strings.Contains(output, "Require stack:")
}

// nodeMissingModuleNewCommand installs the missing npm package and reruns the
// failed command. Relative paths and files are not npm packages, so they are
// left alone.
func nodeMissingModuleNewCommand(command *types.Command) ([]string, error) {
	match := nodeMissingModule.FindStringSubmatch(command.Output)
	if match == nil {
		return nil, nil
	}
	module := match[1]
	if !nodePackageName(module) {
		return nil, nil
	}
	fixed := shells.Current().And("npm install "+module, command.Script)
	return []string{fixed}, nil
}

// nodePackageName reports whether the module name looks like an npm package
// and not like a file or directory path.
func nodePackageName(module string) bool {
	if module == "" {
		return false
	}
	if strings.HasPrefix(module, ".") {
		return false
	}
	if strings.ContainsAny(module, "\\:") {
		return false
	}
	for _, suffix := range []string{".js", ".json", ".mjs", ".cjs", ".ts"} {
		if strings.HasSuffix(module, suffix) {
			return false
		}
	}
	return true
}

func init() {
	register(types.NewRule(
		"node_missing_module",
		utils.ForApp([]string{"node"}, 0, nodeMissingModuleMatch),
		nodeMissingModuleNewCommand,
	))
}
