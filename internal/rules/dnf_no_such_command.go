package rules

import (
	"os/exec"
	"regexp"
	"strings"

	"github.com/leafevans/go-fuck/internal/specific"
	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// dnfNoSuchCommandRegexp matches the misspelled command in the dnf output.
var dnfNoSuchCommandRegexp = regexp.MustCompile(`No such command: (.*)\.`)

// dnfNoSuchCommandOperationsRegexp matches the command names in the dnf help
// output.
var dnfNoSuchCommandOperationsRegexp = regexp.MustCompile(`(?m)^([a-z-]+) +`)

// dnfNoSuchCommandHelp returns the output of "dnf --help".
var dnfNoSuchCommandHelp = utils.Memoize(func() string {
	output, err := exec.Command("dnf", "--help").Output()
	if err != nil {
		return ""
	}
	return string(output)
})

// dnfNoSuchCommandParseOperations parses the operations from the dnf help text.
func dnfNoSuchCommandParseOperations(help string) []string {
	matches := dnfNoSuchCommandOperationsRegexp.FindAllStringSubmatch(help, -1)
	operations := make([]string, 0, len(matches))
	for _, match := range matches {
		operations = append(operations, match[1])
	}
	return operations
}

// dnfNoSuchCommandMatch reports whether dnf rejected an unknown command.
func dnfNoSuchCommandMatch(command *types.Command) bool {
	return strings.Contains(strings.ToLower(command.Output), "no such command")
}

// dnfNoSuchCommandNewCommand suggests the closest valid dnf commands.
func dnfNoSuchCommandNewCommand(command *types.Command) ([]string, error) {
	matches := dnfNoSuchCommandRegexp.FindAllStringSubmatch(command.Output, -1)
	if len(matches) == 0 {
		return nil, nil
	}
	operations := dnfNoSuchCommandParseOperations(dnfNoSuchCommandHelp())
	return utils.ReplaceCommand(command, matches[0][1], operations), nil
}

func init() {
	rule := types.NewRule("dnf_no_such_command",
		specific.SudoSupportMatch(utils.ForApp([]string{"dnf"}, 0, dnfNoSuchCommandMatch)),
		specific.SudoSupportNew(dnfNoSuchCommandNewCommand))
	rule.EnabledByDefault = specific.DnfAvailable()
	register(rule)
}
