package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/internal/shells"
	"github.com/leafevans/go-fuck/internal/specific"
	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// npmCILockMismatchError matches the npm error for a lock file that is out of
// sync with package.json.
const npmCILockMismatchError = "can only install packages when your package.json and package-lock.json"

// npmCILockMismatchMatch reports whether npm ci failed because package-lock.json
// is out of sync.
func npmCILockMismatchMatch(command *types.Command) bool {
	return strings.Contains(command.Output, npmCILockMismatchError)
}

// npmCILockMismatchNewCommand updates the lock file with npm install and then
// retries the clean install.
func npmCILockMismatchNewCommand(command *types.Command) ([]string, error) {
	return []string{shells.Current().And("npm install", command.Script)}, nil
}

func init() {
	// Python: enabled_by_default = npm_available
	rule := types.NewRule("npm_ci_lock_mismatch",
		utils.ForApp([]string{"npm"}, 0, npmCILockMismatchMatch),
		npmCILockMismatchNewCommand)
	rule.EnabledByDefault = specific.NpmAvailable()
	register(rule)
}
