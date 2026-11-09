package rules

import (
	"regexp"

	"github.com/leafevans/go-fuck/internal/shells"
	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// rustupMissingToolchain matches the error of a toolchain that is not
// installed.
var rustupMissingToolchain = regexp.MustCompile(`toolchain '([^']+)' is not installed`)

// rustupToolchainApps are the commands that use the rustup toolchain shim.
var rustupToolchainApps = []string{"rustup", "cargo", "rustc", "rustdoc"}

// rustupToolchainMissingMatch reports whether the command failed because the
// toolchain is not installed.
func rustupToolchainMissingMatch(command *types.Command) bool {
	return utils.IsApp(command, 0, rustupToolchainApps...) &&
		rustupMissingToolchain.MatchString(command.Output)
}

// rustupToolchainMissingNewCommand installs the toolchain and retries.
func rustupToolchainMissingNewCommand(command *types.Command) ([]string, error) {
	match := rustupMissingToolchain.FindStringSubmatch(command.Output)
	if match == nil {
		return nil, nil
	}
	install := "rustup toolchain install " + match[1]
	return []string{shells.Current().And(install, command.Script)}, nil
}

func init() {
	register(types.NewRule("rustup_toolchain_missing",
		rustupToolchainMissingMatch,
		rustupToolchainMissingNewCommand))
}
