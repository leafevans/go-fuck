package rules

import (
	"os/exec"
	"strings"

	"github.com/leafevans/go-fuck/internal/specific"
	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// aptInvalidOperationHelp returns the --help output of the given apt command.
var aptInvalidOperationHelp = utils.Memoize1(func(app string) string {
	output, err := exec.Command(app, "--help").Output()
	if err != nil {
		return ""
	}
	return string(output)
})

// aptInvalidOperationParseAptOperations parses the operation list of apt from
// its help output.
func aptInvalidOperationParseAptOperations(help string) []string {
	operations := []string{}
	isCommandsList := false
	for _, line := range strings.Split(help, "\n") {
		line = strings.TrimSpace(line)
		if isCommandsList && line != "" {
			operations = append(operations, strings.Fields(line)[0])
		} else if strings.HasPrefix(line, "Basic commands:") ||
			strings.HasPrefix(line, "Most used commands:") {
			isCommandsList = true
		}
	}
	return operations
}

// aptInvalidOperationParseAptGetOperations parses the operation list of
// apt-get and apt-cache from their help output.
func aptInvalidOperationParseAptGetOperations(help string) []string {
	operations := []string{}
	isCommandsList := false
	for _, line := range strings.Split(help, "\n") {
		line = strings.TrimSpace(line)
		if isCommandsList {
			if line == "" {
				return operations
			}
			operations = append(operations, strings.Fields(line)[0])
		} else if strings.HasPrefix(line, "Commands:") ||
			strings.HasPrefix(line, "Most used commands:") {
			isCommandsList = true
		}
	}
	return operations
}

// aptInvalidOperationOperations returns the operations supported by the apt
// command.
func aptInvalidOperationOperations(app string) []string {
	help := aptInvalidOperationHelp(app)
	if app == "apt" {
		return aptInvalidOperationParseAptOperations(help)
	}
	return aptInvalidOperationParseAptGetOperations(help)
}

// aptInvalidOperationMatch reports whether apt rejected an unknown operation.
func aptInvalidOperationMatch(command *types.Command) bool {
	return strings.Contains(command.Output, "E: Invalid operation")
}

// aptInvalidOperationNewCommand suggests the closest valid operation.
func aptInvalidOperationNewCommand(command *types.Command) ([]string, error) {
	fields := strings.Fields(command.Output)
	if len(fields) == 0 {
		return nil, nil
	}
	invalidOperation := fields[len(fields)-1]
	if invalidOperation == "uninstall" {
		return []string{strings.ReplaceAll(command.Script, "uninstall", "remove")}, nil
	}

	parts := command.ScriptParts()
	if len(parts) == 0 {
		return nil, nil
	}
	return utils.ReplaceCommand(command, invalidOperation, aptInvalidOperationOperations(parts[0])), nil
}

func init() {
	rule := types.NewRule("apt_invalid_operation",
		specific.SudoSupportMatch(utils.ForApp(
			[]string{"apt", "apt-get", "apt-cache"}, 0, aptInvalidOperationMatch)),
		specific.SudoSupportNew(aptInvalidOperationNewCommand))
	rule.EnabledByDefault = specific.AptAvailable()
	register(rule)
}
