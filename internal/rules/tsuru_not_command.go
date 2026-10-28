package rules

import (
	"regexp"
	"strings"

	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// tsuruNotCommandMistake matches the misspelled tsuru command.
var tsuruNotCommandMistake = regexp.MustCompile(`tsuru: "([^"]*)" is not a tsuru command`)

// tsuruNotCommandMatch reports whether tsuru did not recognize a command.
func tsuruNotCommandMatch(command *types.Command) bool {
	return strings.Contains(command.Output, " is not a tsuru command. See \"tsuru help\".") &&
		strings.Contains(command.Output, newline+"Did you mean?"+newline+"\t")
}

// tsuruNotCommandNewCommand suggests the closest tsuru commands.
func tsuruNotCommandNewCommand(command *types.Command) ([]string, error) {
	match := tsuruNotCommandMistake.FindStringSubmatch(command.Output)
	if match == nil {
		return nil, nil
	}
	matched := utils.GetAllMatchedCommands(command.Output, "Did you mean")
	return utils.ReplaceCommand(command, match[1], matched), nil
}

func init() {
	register(types.NewRule("tsuru_not_command",
		utils.ForApp([]string{"tsuru"}, 0, tsuruNotCommandMatch),
		tsuruNotCommandNewCommand))
}
