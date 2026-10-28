package rules

import (
	"regexp"
	"sort"
	"strings"

	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// tmuxAmbiguous matches the ambiguous command error message of tmux.
var tmuxAmbiguous = regexp.MustCompile(`ambiguous command: (.*), could be: (.*)`)

// tmuxMatch reports whether tmux could not resolve an ambiguous command.
func tmuxMatch(command *types.Command) bool {
	return strings.Contains(command.Output, "ambiguous command:") &&
		strings.Contains(command.Output, "could be:")
}

// tmuxNewCommand suggests the closest tmux commands.
func tmuxNewCommand(command *types.Command) ([]string, error) {
	match := tmuxAmbiguous.FindStringSubmatch(command.Output)
	if match == nil {
		return nil, nil
	}

	suggestions := []string{}
	for _, suggestion := range strings.Split(match[2], ",") {
		suggestions = append(suggestions, strings.TrimSpace(suggestion))
	}
	// Python's difflib orders equally similar suggestions by the suggestion
	// itself. utils.ReplaceCommand keeps the input order for equal ratios, so
	// sort the suggestions the same way beforehand.
	sort.SliceStable(suggestions, func(left, right int) bool {
		return suggestions[left] > suggestions[right]
	})
	return utils.ReplaceCommand(command, match[1], suggestions), nil
}

func init() {
	register(types.NewRule("tmux",
		utils.ForApp([]string{"tmux"}, 0, tmuxMatch),
		tmuxNewCommand))
}
