package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/v2/internal/shells"
	"github.com/leafevans/go-fuck/v2/internal/types"
	"github.com/leafevans/go-fuck/v2/internal/utils"
)

// cargoMissingManifestError is the error cargo prints when no manifest exists.
const cargoMissingManifestError = "could not find `Cargo.toml` in `"

// cargoMissingManifestMatch reports whether cargo failed because no manifest
// exists.
func cargoMissingManifestMatch(command *types.Command) bool {
	return strings.Contains(command.Output, cargoMissingManifestError)
}

// cargoMissingManifestNewCommand initializes a package and retries.
func cargoMissingManifestNewCommand(command *types.Command) ([]string, error) {
	return []string{shells.Current().And("cargo init", command.Script)}, nil
}

func init() {
	register(types.NewRule("cargo_missing_manifest",
		utils.ForApp([]string{"cargo"}, 0, cargoMissingManifestMatch),
		cargoMissingManifestNewCommand))
}
