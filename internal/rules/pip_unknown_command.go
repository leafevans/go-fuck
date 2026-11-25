package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/v4/internal/specific"
	"github.com/leafevans/go-fuck/v4/internal/types"
	"github.com/leafevans/go-fuck/v4/internal/utils"
)

// pipUnknownCommandBroken matches the unknown pip command in the output.
var pipUnknownCommandBroken = newRegexp(`ERROR: unknown command "([^"]+)"`)

// pipUnknownCommandSuggestion matches the pip suggestion in the output.
var pipUnknownCommandSuggestion = newRegexp(`maybe you meant "([^"]+)"`)

// pipUnknownCommandMatch reports whether pip reported an unknown command
// together with a suggestion.
func pipUnknownCommandMatch(command *types.Command) bool {
	return strings.Contains(command.Script, "pip") &&
		strings.Contains(command.Output, "unknown command") &&
		strings.Contains(command.Output, "maybe you meant")
}

func pipGlobalOption(name string) commandOption {
	switch name {
	case "--python", "--log", "--proxy", "--retries", "--timeout", "--trusted-host", "--cert", "--client-cert", "--cache-dir":
		return valueOption
	case "--isolated", "--require-virtualenv", "-v", "-vv", "-vvv", "--verbose", "-q", "--quiet", "--no-cache-dir", "--disable-pip-version-check":
		return flagOption
	default:
		return unknownOption
	}
}

// pipUnknownCommandNewCommand replaces the unknown pip command with the
// suggested one.
func pipUnknownCommandNewCommand(command *types.Command) ([]string, error) {
	broken := pipUnknownCommandBroken.FindStringSubmatch(command.Output)
	suggested := pipUnknownCommandSuggestion.FindStringSubmatch(command.Output)
	if broken == nil || suggested == nil {
		return nil, nil
	}
	words := simpleCommandWords(command.Script)
	return replaceSubcommand(command.Script, words, subcommandIndex(words, 1, pipGlobalOption), broken[1], suggested[1]), nil
}

func init() {
	register(types.NewRule("pip_unknown_command",
		specific.SudoSupportMatch(utils.ForApp([]string{"pip", "pip2", "pip3"}, 0, pipUnknownCommandMatch)),
		specific.SudoSupportNew(pipUnknownCommandNewCommand)))
}
