package rules

import (
	"regexp"
	"strings"

	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// sudoCommandFromUserPathPattern matches the command name reported by sudo.
var sudoCommandFromUserPathPattern = regexp.MustCompile(`sudo: (.*): command not found`)

// sudoCommandFromUserPathWhich reports whether a command is available. It is a
// variable so that tests can inject a fixed result.
var sudoCommandFromUserPathWhich = utils.Which

// sudoCommandFromUserPathCommandName returns the command sudo could not find.
func sudoCommandFromUserPathCommandName(command *types.Command) string {
	found := sudoCommandFromUserPathPattern.FindAllStringSubmatch(command.Output, -1)
	if len(found) == 0 {
		return ""
	}
	return found[0][1]
}

// sudoCommandFromUserPathMatch reports whether sudo could not find a command
// that is available in the user PATH.
func sudoCommandFromUserPathMatch(command *types.Command) bool {
	if !strings.Contains(command.Output, "command not found") {
		return false
	}
	name := sudoCommandFromUserPathCommandName(command)
	return name != "" && sudoCommandFromUserPathWhich(name) != ""
}

// sudoCommandFromUserPathNewCommand passes the user PATH to sudo.
func sudoCommandFromUserPathNewCommand(command *types.Command) ([]string, error) {
	name := sudoCommandFromUserPathCommandName(command)
	if name == "" {
		return nil, nil
	}
	replacement := `env "PATH=$PATH" ` + name
	return []string{utils.ReplaceArgument(command.Script, name, replacement)}, nil
}

func init() {
	register(types.NewRule("sudo_command_from_user_path",
		utils.ForApp([]string{"sudo"}, 0, sudoCommandFromUserPathMatch),
		sudoCommandFromUserPathNewCommand))
}
