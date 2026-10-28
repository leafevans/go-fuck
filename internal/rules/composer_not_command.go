package rules

import (
	"regexp"
	"strings"

	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// composerNotCommandBrokenRegexp matches the misspelled composer command.
var composerNotCommandBrokenRegexp = regexp.MustCompile(`Command "([^']*)" is not defined`)

// composerNotCommandThisRegexp matches a single suggestion.
var composerNotCommandThisRegexp = regexp.MustCompile(
	`Did you mean this\?[^\n]*\n\s*([^\n]*)`)

// composerNotCommandOneOfThisRegexp matches a list of suggestions.
var composerNotCommandOneOfThisRegexp = regexp.MustCompile(
	`Did you mean one of these\?[^\n]*\n\s*([^\n]*)`)

// composerNotCommandMatch reports whether composer did not recognize the
// command or package.
func composerNotCommandMatch(command *types.Command) bool {
	lower := strings.ToLower(command.Output)
	if strings.Contains(lower, "did you mean this?") ||
		strings.Contains(lower, "did you mean one of these?") {
		return true
	}
	return utils.Contains(command.ScriptParts(), "install") &&
		strings.Contains(lower, "composer require")
}

// composerNotCommandNewCommand suggests the closest valid command.
func composerNotCommandNewCommand(command *types.Command) ([]string, error) {
	lower := strings.ToLower(command.Output)
	brokenCommand, newCommand := "", ""
	if utils.Contains(command.ScriptParts(), "install") &&
		strings.Contains(lower, "composer require") {
		brokenCommand, newCommand = "install", "require"
	} else {
		brokenMatch := composerNotCommandBrokenRegexp.FindStringSubmatch(command.Output)
		if brokenMatch == nil {
			return nil, nil
		}
		brokenCommand = brokenMatch[1]
		if match := composerNotCommandThisRegexp.FindStringSubmatch(command.Output); match != nil {
			newCommand = strings.TrimSpace(match[1])
		} else if match := composerNotCommandOneOfThisRegexp.FindStringSubmatch(command.Output); match != nil {
			newCommand = strings.TrimSpace(match[1])
		} else {
			return nil, nil
		}
	}
	return []string{utils.ReplaceArgument(command.Script, brokenCommand, newCommand)}, nil
}

func init() {
	register(types.NewRule("composer_not_command",
		utils.ForApp([]string{"composer"}, 0, composerNotCommandMatch),
		composerNotCommandNewCommand))
}
