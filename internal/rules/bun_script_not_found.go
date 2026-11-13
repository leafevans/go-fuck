package rules

import (
	"regexp"

	"github.com/leafevans/go-fuck/v2/internal/specific"
	"github.com/leafevans/go-fuck/v2/internal/types"
	"github.com/leafevans/go-fuck/v2/internal/utils"
)

var (
	// bunScriptNotFound matches the error line of bun for a script that does
	// not exist, for example: error: Script not found "buidl".
	bunScriptNotFound = regexp.MustCompile(`Script not found "([^"]*)"`)
	// bunAnsiColor matches ANSI SGR sequences around the colored bun error.
	bunAnsiColor = regexp.MustCompile(`\x1b\[[0-9;]*m`)
)

// bunOutput strips ANSI color sequences from the captured bun output.
func bunOutput(command *types.Command) string {
	return bunAnsiColor.ReplaceAllString(command.Output, "")
}

// bunScriptNotFoundMatch reports whether bun failed because a package script
// does not exist.
func bunScriptNotFoundMatch(command *types.Command) bool {
	if bunScriptNotFound.FindStringSubmatch(bunOutput(command)) == nil {
		return false
	}
	return len(specific.PackageScripts()) > 0
}

// bunScriptNotFoundNewCommand suggests the package.json scripts that are close
// to the misspelled one.
func bunScriptNotFoundNewCommand(command *types.Command) ([]string, error) {
	match := bunScriptNotFound.FindStringSubmatch(bunOutput(command))
	if match == nil {
		return nil, nil
	}
	return utils.ReplaceCommand(command, match[1], specific.PackageScripts()), nil
}

func init() {
	register(types.NewRule(
		"bun_script_not_found",
		utils.ForApp([]string{"bun"}, 0, bunScriptNotFoundMatch),
		bunScriptNotFoundNewCommand,
	))
}
