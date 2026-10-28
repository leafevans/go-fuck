package rules

import (
	"os/exec"
	"strings"

	"github.com/leafevans/go-fuck/internal/specific"
	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// yumInvalidOperationHelp returns the output of yum run without arguments.
var yumInvalidOperationHelp = utils.Memoize(func() string {
	output, err := exec.Command("yum").Output()
	if err != nil {
		return ""
	}
	return string(output)
})

// yumInvalidOperationParseOperations parses the command list of the yum help
// output.
func yumInvalidOperationParseOperations(help string) []string {
	lines := strings.Split(help, "\n")
	start := -1
	for index, line := range lines {
		if strings.HasPrefix(line, "List of Commands:") {
			start = index + 2
			break
		}
	}
	if start < 0 || start > len(lines) {
		return nil
	}

	operations := []string{}
	for _, line := range lines[start:] {
		line = strings.TrimSpace(line)
		if line == "" {
			break
		}
		operations = append(operations, strings.Fields(line)[0])
	}
	return operations
}

// yumInvalidOperationMatch reports whether yum rejected an unknown operation.
func yumInvalidOperationMatch(command *types.Command) bool {
	return strings.Contains(command.Output, "No such command: ")
}

// yumInvalidOperationNewCommand suggests the closest valid operation.
func yumInvalidOperationNewCommand(command *types.Command) ([]string, error) {
	parts := command.ScriptParts()
	if len(parts) < 2 {
		return nil, nil
	}
	invalidOperation := parts[1]
	if invalidOperation == "uninstall" {
		return []string{strings.ReplaceAll(command.Script, "uninstall", "remove")}, nil
	}
	operations := yumInvalidOperationParseOperations(yumInvalidOperationHelp())
	return utils.ReplaceCommand(command, invalidOperation, operations), nil
}

func init() {
	rule := types.NewRule("yum_invalid_operation",
		specific.SudoSupportMatch(utils.ForApp([]string{"yum"}, 0, yumInvalidOperationMatch)),
		specific.SudoSupportNew(yumInvalidOperationNewCommand))
	rule.EnabledByDefault = specific.YumAvailable()
	register(rule)
}
