package rules

import (
	"os/exec"
	"strings"

	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// ifconfigDeviceNotFoundError is the error printed for a missing interface.
const ifconfigDeviceNotFoundError = "error fetching interface information: Device not found"

// ifconfigReadOutput runs "ifconfig -a" and returns its output.
var ifconfigReadOutput = func() string {
	output, _ := exec.Command("ifconfig", "-a").Output()
	return string(output)
}

// ifconfigParseInterfaces extracts the interface names from "ifconfig -a".
func ifconfigParseInterfaces(output string) []string {
	interfaces := []string{}
	for _, line := range strings.Split(output, newline) {
		// Interface headers start at the beginning of the line, the details
		// of an interface are indented.
		if line == "" || strings.HasPrefix(line, " ") {
			continue
		}
		interfaces = append(interfaces, strings.Split(line, " ")[0])
	}
	return interfaces
}

// ifconfigPossibleInterfaces returns the available network interfaces.
var ifconfigPossibleInterfaces = utils.Memoize(func() []string {
	return ifconfigParseInterfaces(ifconfigReadOutput())
})

// ifconfigDeviceNotFoundMatch reports whether ifconfig could not find the
// device.
func ifconfigDeviceNotFoundMatch(command *types.Command) bool {
	return strings.Contains(command.Output, ifconfigDeviceNotFoundError)
}

// ifconfigDeviceNotFoundNewCommand suggests the closest interface name.
func ifconfigDeviceNotFoundNewCommand(command *types.Command) ([]string, error) {
	// Python: command.output.split(' ')[0][:-1] -- drop the trailing colon.
	first := strings.SplitN(command.Output, " ", 2)[0]
	runes := []rune(first)
	if len(runes) < 2 {
		return nil, nil
	}
	interfaceName := string(runes[:len(runes)-1])
	return utils.ReplaceCommand(command, interfaceName, ifconfigPossibleInterfaces()), nil
}

func init() {
	register(types.NewRule("ifconfig_device_not_found",
		utils.ForApp([]string{"ifconfig"}, 0, ifconfigDeviceNotFoundMatch),
		ifconfigDeviceNotFoundNewCommand))
}
