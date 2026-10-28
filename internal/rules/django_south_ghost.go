package rules

import (
	"strings"

	"github.com/leafevans/go-fuck/internal/types"
)

// djangoSouthGhostMatch reports whether django south found ghost migrations.
func djangoSouthGhostMatch(command *types.Command) bool {
	return strings.Contains(command.Script, "manage.py") &&
		strings.Contains(command.Script, "migrate") &&
		strings.Contains(command.Output, "or pass --delete-ghost-migrations")
}

// djangoSouthGhostNewCommand deletes the ghost migrations.
func djangoSouthGhostNewCommand(command *types.Command) ([]string, error) {
	return []string{command.Script + " --delete-ghost-migrations"}, nil
}

func init() {
	register(types.NewRule("django_south_ghost",
		djangoSouthGhostMatch,
		djangoSouthGhostNewCommand))
}
