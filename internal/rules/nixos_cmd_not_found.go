package rules

import (
	"regexp"

	"github.com/leafevans/go-fuck/internal/shells"
	"github.com/leafevans/go-fuck/internal/specific"
	"github.com/leafevans/go-fuck/internal/types"
)

// nixosCmdNotFoundRegexp matches the package installation hint printed by
// NixOS.
var nixosCmdNotFoundRegexp = regexp.MustCompile(`nix-env -iA ([^\s]*)`)

// nixosCmdNotFoundMatch reports whether the shell printed the nix-env hint.
func nixosCmdNotFoundMatch(command *types.Command) bool {
	return len(nixosCmdNotFoundRegexp.FindAllStringSubmatch(command.Output, -1)) > 0
}

// nixosCmdNotFoundNewCommand installs the missing package and re-runs the
// command.
func nixosCmdNotFoundNewCommand(command *types.Command) ([]string, error) {
	matches := nixosCmdNotFoundRegexp.FindAllStringSubmatch(command.Output, -1)
	if len(matches) == 0 {
		return nil, nil
	}
	return []string{shells.Current().And("nix-env -iA "+matches[0][1], command.Script)}, nil
}

func init() {
	rule := types.NewRule("nixos_cmd_not_found", nixosCmdNotFoundMatch, nixosCmdNotFoundNewCommand)
	rule.EnabledByDefault = specific.NixAvailable()
	register(rule)
}
