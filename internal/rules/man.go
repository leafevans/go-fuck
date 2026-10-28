package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// manMatch reports whether the command calls man with an argument. The
// application check and the minimum number of arguments are done by the
// for_app wrapper.
func manMatch(*types.Command) bool {
	return true
}

// manNewCommand suggests the other man page sections and the --help flag.
func manNewCommand(command *types.Command) ([]string, error) {
	if strings.Contains(command.Script, "3") {
		return []string{strings.ReplaceAll(command.Script, "3", "2")}, nil
	}
	if strings.Contains(command.Script, "2") {
		return []string{strings.ReplaceAll(command.Script, "2", "3")}, nil
	}

	parts := command.ScriptParts()
	if len(parts) == 0 {
		return nil, nil
	}
	lastArgument := parts[len(parts)-1]
	helpCommand := lastArgument + " --help"

	// If there are no man pages for the argument, suggest its --help flag
	// instead of the other man page sections.
	if strings.TrimSpace(command.Output) == "No manual entry for "+lastArgument {
		return []string{helpCommand}, nil
	}

	// The reference implementation inserts the section as a separate list
	// element that already contains the surrounding spaces and then joins the
	// parts without a separator.
	withSection := func(section string) string {
		inserted := make([]string, 0, len(parts)+1)
		inserted = append(inserted, parts[0])
		inserted = append(inserted, " "+section+" ")
		inserted = append(inserted, parts[1:]...)
		return strings.Join(inserted, "")
	}

	return []string{withSection("3"), withSection("2"), helpCommand}, nil
}

func init() {
	register(types.NewRule("man", utils.ForApp([]string{"man"}, 1, manMatch), manNewCommand))
}
