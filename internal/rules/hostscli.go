package rules

import (
	"regexp"
	"strings"

	"github.com/leafevans/go-fuck/internal/specific"
	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// hostscliNoCommand is the error hostscli prints for an unknown command.
const hostscliNoCommand = "Error: No such command"

// hostscliNoWebsite is the error hostscli prints for an unknown website.
const hostscliNoWebsite = "hostscli.errors.WebsiteImportError"

// hostscliMisspelledCommand matches the unknown command reported by hostscli.
var hostscliMisspelledCommand = regexp.MustCompile(`Error: No such command ".*"`)

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

	matches := hostscliMisspelledCommand.FindAllString(command.Output, -1)
	if len(matches) == 0 {
		return nil, nil
	}
	return utils.ReplaceCommand(command, matches[0], hostscliCommands), nil
}

func init() {
	register(types.NewRule("hostscli",
		specific.SudoSupportMatch(utils.ForApp([]string{"hostscli"}, 0, hostscliMatch)),
		specific.SudoSupportNew(hostscliNewCommand)))
}
