package rules

import (
	"regexp"
	"strings"

	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// longFormHelpSuggestion matches the message that suggests the long form of
// the help flag.
var longFormHelpSuggestion = regexp.MustCompile(`(?:Run|Try) '([^']+)'(?: or '[^']+')? for (?:details|more information).`)

// longFormHelpSuggestionInsensitive is the case insensitive variant, used to
// extract the suggested help command.
var longFormHelpSuggestionInsensitive = regexp.MustCompile(`(?i)(?:Run|Try) '([^']+)'(?: or '[^']+')? for (?:details|more information).`)

// longFormHelpMatch reports whether the output suggests a help command or
// mentions the long help flag.
func longFormHelpMatch(command *types.Command) bool {
	return longFormHelpSuggestionInsensitive.MatchString(command.Output) ||
		strings.Contains(command.Output, "--help")
}

// longFormHelpNewCommand suggests the command printed by the tool, or replaces
// the short help flag with the long one.
func longFormHelpNewCommand(command *types.Command) ([]string, error) {
	if longFormHelpSuggestion.MatchString(command.Output) {
		match := longFormHelpSuggestionInsensitive.FindStringSubmatch(command.Output)
		if match != nil {
			return []string{match[1]}, nil
		}
	}
	return []string{utils.ReplaceArgument(command.Script, "-h", "--help")}, nil
}

func init() {
	rule := types.NewRule("long_form_help", longFormHelpMatch, longFormHelpNewCommand)
	rule.EnabledByDefault = true
	rule.Priority = 5000
	register(rule)
}
