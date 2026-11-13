package rules

import (
	"regexp"
	"strings"

	"github.com/leafevans/go-fuck/v2/internal/specific"
	"github.com/leafevans/go-fuck/v2/internal/types"
	"github.com/leafevans/go-fuck/v2/internal/utils"
)

// npmMissingScriptPattern extracts the misspelled script name from npm
// output. npm 6 and older print "npm ERR! missing script: wach"; npm 7 and
// newer print 'npm error Missing script: "wach"'.
var npmMissingScriptPattern = regexp.MustCompile(`(?i)missing script: "?([^"\n]+)"?`)

// npmMissingScriptRunPartPrefix is the prefix of the npm arguments that run a
// script ("run", "ru", "run-script", ...).
const npmMissingScriptRunPartPrefix = "ru"

// npmMissingScriptMatch reports whether npm failed because a script is missing.
func npmMissingScriptMatch(command *types.Command) bool {
	hasRunPart := false
	for _, part := range command.ScriptParts() {
		if strings.HasPrefix(part, npmMissingScriptRunPartPrefix) {
			hasRunPart = true
			break
		}
	}
	return hasRunPart && npmMissingScriptPattern.MatchString(command.Output)
}

// npmMissingScriptNewCommand suggests the npm scripts that are close to the
// misspelled one.
func npmMissingScriptNewCommand(command *types.Command) ([]string, error) {
	match := npmMissingScriptPattern.FindStringSubmatch(command.Output)
	if match == nil {
		return nil, nil
	}
	return utils.ReplaceCommand(command, match[1], specific.NpmGetScripts()), nil
}

func init() {
	// Python: enabled_by_default = npm_available
	rule := types.NewRule("npm_missing_script",
		utils.ForApp([]string{"npm"}, 0, npmMissingScriptMatch),
		npmMissingScriptNewCommand)
	rule.EnabledByDefault = specific.NpmAvailable()
	register(rule)
}
