package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/internal/shells"
	"github.com/leafevans/go-fuck/internal/specific"
	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// npmCIMissingLockfileError matches the npm error for a clean install without
// a lock file.
const npmCIMissingLockfileError = "can only install with an existing package-lock.json"

// npmCIMissingLockfileMatch reports whether npm ci failed because there is no
// package-lock.json.
func npmCIMissingLockfileMatch(command *types.Command) bool {
	return strings.Contains(command.Output, npmCIMissingLockfileError)
}

// npmCIMissingLockfileNewCommand generates the lock file with npm install and
// then retries the clean install.
func npmCIMissingLockfileNewCommand(command *types.Command) ([]string, error) {
	return []string{shells.Current().And("npm install", command.Script)}, nil
}

func init() {
	// Python: enabled_by_default = npm_available
	rule := types.NewRule("npm_ci_missing_lockfile",
		utils.ForApp([]string{"npm"}, 0, npmCIMissingLockfileMatch),
		npmCIMissingLockfileNewCommand)
	rule.EnabledByDefault = specific.NpmAvailable()
	register(rule)
}
