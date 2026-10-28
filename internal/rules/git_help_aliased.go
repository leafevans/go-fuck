package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/internal/specific"
	"github.com/leafevans/go-fuck/internal/types"
)

// gitHelpAliasedMatch reports whether git help was called for an alias.
func gitHelpAliasedMatch(command *types.Command) bool {
	return strings.Contains(command.Script, "help") &&
		strings.Contains(command.Output, " is aliased to ")
}

// gitHelpAliasedNewCommand asks for the help of the aliased command.
func gitHelpAliasedNewCommand(command *types.Command) ([]string, error) {
	backtick := string(rune(96))
	output := command.Output

	index := strings.Index(output, backtick)
	if index < 0 {
		return nil, nil
	}
	rest := output[index+1:]
	index = strings.Index(rest, backtick)
	if index < 0 {
		return nil, nil
	}
	rest = rest[index+1:]

	aliased := strings.SplitN(rest, "'", 2)[0]
	aliased = strings.SplitN(aliased, " ", 2)[0]
	return []string{"git help " + aliased}, nil
}

func init() {
	register(types.NewRule("git_help_aliased",
		specific.GitSupportMatch(gitHelpAliasedMatch),
		specific.GitSupportNew(gitHelpAliasedNewCommand)))
}
