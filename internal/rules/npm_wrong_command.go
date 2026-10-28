package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/internal/specific"
	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// npmWrongCommandListing is the line that starts the npm command listing in the
// npm help output.
const npmWrongCommandListing = "where <command> is one of:"

// npmWrongCommandFirst returns the first non-flag argument of the command.
func npmWrongCommandFirst(parts []string) string {
	for _, part := range parts[1:] {
		if !strings.HasPrefix(part, "-") {
			return part
		}
	}
	return ""
}

// npmWrongCommandMatch reports whether an unknown npm command was used.
func npmWrongCommandMatch(command *types.Command) bool {
	parts := command.ScriptParts()
	if len(parts) == 0 {
		return false
	}
	return parts[0] == "npm" &&
		strings.Contains(command.Output, npmWrongCommandListing) &&
		npmWrongCommandFirst(parts) != ""
}

// npmWrongCommandAvailable returns the npm commands listed in the help output.
func npmWrongCommandAvailable(commandOutput string) []string {
	commands := []string{}
	listing := false
	for _, line := range strings.Split(commandOutput, "\n") {
		if strings.HasPrefix(line, npmWrongCommandListing) {
			listing = true
			continue
		}
		if !listing {
			continue
		}
		if line == "" {
			break
		}
		for _, command := range strings.Split(line, ", ") {
			if trimmed := strings.TrimSpace(command); trimmed != "" {
				commands = append(commands, trimmed)
			}
		}
	}
	return commands
}

// npmWrongCommandNewCommand replaces the unknown command with the closest npm
// command.
func npmWrongCommandNewCommand(command *types.Command) ([]string, error) {
	wrongCommand := npmWrongCommandFirst(command.ScriptParts())
	if wrongCommand == "" {
		return nil, nil
	}
	fixed := utils.GetClosestDefault(wrongCommand, npmWrongCommandAvailable(command.Output))
	if fixed == "" {
		return nil, nil
	}
	return []string{utils.ReplaceArgument(command.Script, wrongCommand, fixed)}, nil
}

func init() {
	// Python: enabled_by_default = npm_available
	rule := types.NewRule("npm_wrong_command",
		specific.SudoSupportMatch(utils.ForApp([]string{"npm"}, 0, npmWrongCommandMatch)),
		npmWrongCommandNewCommand)
	rule.EnabledByDefault = specific.NpmAvailable()
	register(rule)
}
