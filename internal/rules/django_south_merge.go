package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/internal/types"
)

// djangoSouthMergeMatch reports whether django south detected an inconsistent
// migration history.
func djangoSouthMergeMatch(command *types.Command) bool {
	return strings.Contains(command.Script, "manage.py") &&
		strings.Contains(command.Script, "migrate") &&
		strings.Contains(command.Output, "--merge: will just attempt the migration")
}

// djangoSouthMergeNewCommand retries the migration with --merge.
func djangoSouthMergeNewCommand(command *types.Command) ([]string, error) {
	return []string{command.Script + " --merge"}, nil
}

func init() {
	register(types.NewRule("django_south_merge",
		djangoSouthMergeMatch,
		djangoSouthMergeNewCommand))
}
