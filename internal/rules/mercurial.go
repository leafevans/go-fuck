package rules

import (
	"regexp"
	"strings"

	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// mercurialDidYouMean matches the suggestions printed for an unknown command.
var mercurialDidYouMean = regexp.MustCompile(`\n\(did you mean one of ([^?]+)\?\)`)

// mercurialAmbiguous matches the commands listed for an ambiguous command.
var mercurialAmbiguous = regexp.MustCompile(`\n    ([^$]+)$`)

// mercurialExtractPossibilities returns the commands suggested by mercurial.
func mercurialExtractPossibilities(command *types.Command) []string {
	matches := mercurialDidYouMean.FindAllStringSubmatch(command.Output, -1)
	if len(matches) > 0 {
		return strings.Split(matches[0][1], ", ")
	}
	matches = mercurialAmbiguous.FindAllStringSubmatch(command.Output, -1)
	if len(matches) > 0 {
		return strings.Split(matches[0][1], " ")
	}
	return nil
}

// mercurialMatch reports whether mercurial did not recognize a command.
func mercurialMatch(command *types.Command) bool {
	if strings.Contains(command.Output, "hg: unknown command") &&
		strings.Contains(command.Output, "(did you mean one of ") {
		return true
	}
	return strings.Contains(command.Output, "hg: command '") &&
		strings.Contains(command.Output, "' is ambiguous:")
}

// mercurialNewCommand replaces the unknown command with the closest match.
func mercurialNewCommand(command *types.Command) ([]string, error) {
	parts := command.ScriptParts()
	if len(parts) < 2 {
		return nil, nil
	}
	possibilities := mercurialExtractPossibilities(command)
	if len(possibilities) == 0 {
		return nil, nil
	}
	parts = append([]string{}, parts...)
	parts[1] = utils.GetClosestDefault(parts[1], possibilities)
	return []string{strings.Join(parts, " ")}, nil
}

func init() {
	register(types.NewRule("mercurial",
		utils.ForApp([]string{"hg"}, 0, mercurialMatch),
		mercurialNewCommand))
}
