package rules

import (
	"regexp"
	"strings"

	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// herokuNotCommandSuggestion matches the corrected command suggested by
// heroku.
var herokuNotCommandSuggestion = regexp.MustCompile(`Run heroku _ to run ([^.]*)`)

// herokuNotCommandMatch reports whether heroku did not recognize a command.
func herokuNotCommandMatch(command *types.Command) bool {
	return strings.Contains(command.Output, "Run heroku _ to run")
}

// herokuNotCommandNewCommand uses the command suggested by heroku.
func herokuNotCommandNewCommand(command *types.Command) ([]string, error) {
	match := herokuNotCommandSuggestion.FindStringSubmatch(command.Output)
	if match == nil {
		return nil, nil
	}
	return []string{match[1]}, nil
}

func init() {
	register(types.NewRule("heroku_not_command",
		utils.ForApp([]string{"heroku"}, 0, herokuNotCommandMatch),
		herokuNotCommandNewCommand))
}
