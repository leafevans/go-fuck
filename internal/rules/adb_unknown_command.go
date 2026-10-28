package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// adbUnknownCommandCommands lists the adb subcommands.
var adbUnknownCommandCommands = []string{
	"backup",
	"bugreport",
	"connect",
	"devices",
	"disable-verity",
	"disconnect",
	"enable-verity",
	"emu",
	"forward",
	"get-devpath",
	"get-serialno",
	"get-state",
	"install",
	"install-multiple",
	"jdwp",
	"keygen",
	"kill-server",
	"logcat",
	"pull",
	"push",
	"reboot",
	"reconnect",
	"restore",
	"reverse",
	"root",
	"run-as",
	"shell",
	"sideload",
	"start-server",
	"sync",
	"tcpip",
	"uninstall",
	"unroot",
	"usb",
	"wait-for",
}

// adbUnknownCommandOptions lists the adb flags that take an additional
// argument, for example "adb -s 111 logcat" or "adb -e logcat".
var adbUnknownCommandOptions = []string{"-s", "-H", "-P", "-L"}

// adbUnknownCommandMatch reports whether adb printed its help because the
// subcommand is misspelled.
func adbUnknownCommandMatch(command *types.Command) bool {
	return utils.IsApp(command, 0, "adb") &&
		strings.HasPrefix(command.Output, "Android Debug Bridge version")
}

// adbUnknownCommandNewCommand replaces the misspelled subcommand with the
// closest adb command.
func adbUnknownCommandNewCommand(command *types.Command) ([]string, error) {
	parts := command.ScriptParts()
	if len(parts) == 0 {
		return nil, nil
	}
	for index, argument := range parts[1:] {
		if strings.HasPrefix(argument, "-") {
			continue
		}
		if utils.Contains(adbUnknownCommandOptions, parts[index]) {
			// The preceding argument is a flag that takes a value.
			continue
		}
		adbCommand := utils.GetClosestDefault(argument, adbUnknownCommandCommands)
		if adbCommand == "" {
			return nil, nil
		}
		return []string{utils.ReplaceArgument(command.Script, argument, adbCommand)}, nil
	}
	return nil, nil
}

func init() {
	register(types.NewRule("adb_unknown_command",
		adbUnknownCommandMatch,
		adbUnknownCommandNewCommand))
}
