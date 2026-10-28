package rules

import (
	"regexp"
	"strings"

	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// herokuMultipleAppsSuggestion matches the "app (remote)" lines that heroku
// prints when several apps are configured.
var herokuMultipleAppsSuggestion = regexp.MustCompile(`([^ ]*) \([^)]*\)`)

// herokuMultipleAppsMatch reports whether heroku could not choose an app.
func herokuMultipleAppsMatch(command *types.Command) bool {
	return strings.Contains(command.Output, "https://devcenter.heroku.com/articles/multiple-environments")
}

// herokuMultipleAppsNewCommand retries the command for every known app.
func herokuMultipleAppsNewCommand(command *types.Command) ([]string, error) {
	matches := herokuMultipleAppsSuggestion.FindAllStringSubmatch(command.Output, -1)
	results := make([]string, 0, len(matches))
	for _, match := range matches {
		results = append(results, command.Script+" --app "+match[1])
	}
	return results, nil
}

func init() {
	register(types.NewRule("heroku_multiple_apps",
		utils.ForApp([]string{"heroku"}, 0, herokuMultipleAppsMatch),
		herokuMultipleAppsNewCommand))
}
