package rules

import (
	"regexp"
	"strings"

	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// cargoNoCommandSuggestionRegexp matches the subcommand suggested by cargo.
var cargoNoCommandSuggestionRegexp = regexp.MustCompile(`Did you mean ` + "`([^`]*)`")

// cargoNoCommandMatch reports whether cargo did not recognize the subcommand.
func cargoNoCommandMatch(command *types.Command) bool {
	return strings.Contains(strings.ToLower(command.Output), "no such subcommand") &&
		strings.Contains(command.Output, "Did you mean")
}

// cargoNoCommandNewCommand replaces the misspelled subcommand.
func cargoNoCommandNewCommand(command *types.Command) ([]string, error) {
	parts := command.ScriptParts()
	match := cargoNoCommandSuggestionRegexp.FindStringSubmatch(command.Output)
	if len(parts) < 2 || match == nil {
		return nil, nil
	}
	return []string{utils.ReplaceArgument(command.Script, parts[1], match[1])}, nil
}

func init() {
	register(types.NewRule("cargo_no_command",
		utils.ForApp([]string{"cargo"}, 1, cargoNoCommandMatch),
		cargoNoCommandNewCommand))
}
