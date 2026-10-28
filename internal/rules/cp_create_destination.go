package rules

import (
	"strings"
	"unicode"

	"github.com/leafevans/go-fuck/internal/shells"
	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// cpCreateDestinationMatch reports whether cp or mv failed because the
// destination directory is missing.
func cpCreateDestinationMatch(command *types.Command) bool {
	output := command.Output
	// Python: output.startswith("cp: directory") and
	// output.rstrip().endswith("does not exist")
	return strings.Contains(output, "No such file or directory") ||
		strings.HasPrefix(output, "cp: directory") &&
			strings.HasSuffix(strings.TrimRightFunc(output, unicode.IsSpace), "does not exist")
}

// cpCreateDestinationNewCommand creates the destination directory first.
func cpCreateDestinationNewCommand(command *types.Command) ([]string, error) {
	parts := command.ScriptParts()
	if len(parts) == 0 {
		return nil, nil
	}
	destination := parts[len(parts)-1]
	return []string{shells.Current().And("mkdir -p "+destination, command.Script)}, nil
}

func init() {
	register(types.NewRule("cp_create_destination",
		utils.ForApp([]string{"cp", "mv"}, 0, cpCreateDestinationMatch),
		cpCreateDestinationNewCommand))
}
