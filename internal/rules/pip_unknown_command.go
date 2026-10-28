package rules

import (
	"regexp"
	"strings"

	"github.com/leafevans/go-fuck/internal/specific"
	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// pipUnknownCommandBroken matches the unknown pip command in the output.
var pipUnknownCommandBroken = regexp.MustCompile(`ERROR: unknown command "([^"]+)"`)

// pipUnknownCommandSuggestion matches the pip suggestion in the output.
var pipUnknownCommandSuggestion = regexp.MustCompile(`maybe you meant "([^"]+)"`)

// pipUnknownCommandMatch reports whether pip reported an unknown command
// together with a suggestion.
func pipUnknownCommandMatch(command *types.Command) bool {
	return strings.Contains(command.Script, "pip") &&
		strings.Contains(command.Output, "unknown command") &&
		strings.Contains(command.Output, "maybe you meant")
}

// pipUnknownCommandNewCommand replaces the unknown pip command with the
// suggested one.
func pipUnknownCommandNewCommand(command *types.Command) ([]string, error) {
	broken := pipUnknownCommandBroken.FindStringSubmatch(command.Output)
	suggested := pipUnknownCommandSuggestion.FindStringSubmatch(command.Output)
	if broken == nil || suggested == nil {
		return nil, nil
	}
	return []string{utils.ReplaceArgument(command.Script, broken[1], suggested[1])}, nil
}

func init() {
	register(types.NewRule("pip_unknown_command",
		specific.SudoSupportMatch(utils.ForApp([]string{"pip", "pip2", "pip3"}, 0, pipUnknownCommandMatch)),
		pipUnknownCommandNewCommand))
}
