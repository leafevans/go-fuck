package rules

import (
	"strings"
	"unicode"

	"github.com/leafevans/go-fuck/v3/internal/shells"
	"github.com/leafevans/go-fuck/v3/internal/types"
	"github.com/leafevans/go-fuck/v3/internal/utils"
)

// cpCreateDestinationMissingPathPattern matches the cp and mv messages that name
// the destination the command could not create.
var cpCreateDestinationMissingPathPattern = newRegexp(
	`cannot create regular file '[^']*'|cannot move '[^']*' to '[^']*'`)

// cpCreateDestinationMatch reports whether cp or mv failed because the
// destination directory is missing.
//
// Deviation: the reference implementation matches any "No such file or
// directory" output, so a missing source file suggested creating a directory
// named after the destination. Only the messages that name a destination the
// command could not create are matched; a missing source is not a destination.
func cpCreateDestinationMatch(command *types.Command) bool {
	output := command.Output
	// Python: output.startswith("cp: directory") and
	// output.rstrip().endswith("does not exist")
	if strings.HasPrefix(output, "cp: directory") &&
		strings.HasSuffix(strings.TrimRightFunc(output, unicode.IsSpace), "does not exist") {
		return true
	}
	return strings.Contains(output, "No such file or directory") &&
		cpCreateDestinationMissingPathPattern.MatchString(output)
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
