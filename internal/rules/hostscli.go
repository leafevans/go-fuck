package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/v3/internal/specific"
	"github.com/leafevans/go-fuck/v3/internal/types"
	"github.com/leafevans/go-fuck/v3/internal/utils"
)

// hostscliNoCommand is the error hostscli prints for an unknown command.
const hostscliNoCommand = "Error: No such command"

// hostscliNoWebsite is the error hostscli prints for an unknown website.
const hostscliNoWebsite = "hostscli.errors.WebsiteImportError"

// hostscliMisspelledCommand matches the unknown command reported by hostscli.
// The first group is the command name itself: the surrounding error message is
// not part of the command line and cannot be replaced there.
var hostscliMisspelledCommand = newRegexp(`Error: No such command "([^"]*)"`)

// hostscliCommands lists the commands implemented by hostscli.
var hostscliCommands = []string{"block", "unblock", "websites", "block_all", "unblock_all"}

// hostscliMatch reports whether hostscli failed because of an unknown command
// or website.
func hostscliMatch(command *types.Command) bool {
	return strings.Contains(command.Output, hostscliNoCommand) ||
		strings.Contains(command.Output, hostscliNoWebsite)
}

// hostscliNewCommand suggests a valid hostscli command.
func hostscliNewCommand(command *types.Command) ([]string, error) {
	if strings.Contains(command.Output, hostscliNoWebsite) {
		return []string{"hostscli websites"}, nil
	}

	match := hostscliMisspelledCommand.FindStringSubmatch(command.Output)
	if match == nil {
		return nil, nil
	}
	return utils.ReplaceCommand(command, match[1], hostscliCommands), nil
}

func init() {
	register(types.NewRule("hostscli",
		specific.SudoSupportMatch(utils.ForApp([]string{"hostscli"}, 0, hostscliMatch)),
		specific.SudoSupportNew(hostscliNewCommand)))
}
