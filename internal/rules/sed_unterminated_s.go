package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/internal/shells"
	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// sedUnterminatedSPrefixes lists the prefixes of the s command expressions.
var sedUnterminatedSPrefixes = []string{"s/", "-es/"}

// sedUnterminatedSMatch reports whether sed reported an unterminated s
// command.
func sedUnterminatedSMatch(command *types.Command) bool {
	return strings.Contains(command.Output, "unterminated `s' command")
}

// sedUnterminatedSNewCommand terminates the unterminated s commands.
func sedUnterminatedSNewCommand(command *types.Command) ([]string, error) {
	parts, err := shells.Split(command.Script)
	if err != nil {
		return nil, err
	}

	for index, part := range parts {
		matched := false
		for _, prefix := range sedUnterminatedSPrefixes {
			if strings.HasPrefix(part, prefix) {
				matched = true
				break
			}
		}
		if matched && !strings.HasSuffix(part, "/") {
			parts[index] = part + "/"
		}
	}

	quoted := make([]string, len(parts))
	for index, part := range parts {
		quoted[index] = shells.Current().Quote(part)
	}
	return []string{strings.Join(quoted, " ")}, nil
}

func init() {
	register(types.NewRule("sed_unterminated_s",
		utils.ForApp([]string{"sed"}, 0, sedUnterminatedSMatch),
		sedUnterminatedSNewCommand))
}
