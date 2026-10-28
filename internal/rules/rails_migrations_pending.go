package rules

import (
	"regexp"
	"strings"

	"github.com/leafevans/go-fuck/internal/shells"
	"github.com/leafevans/go-fuck/internal/types"
)

// railsMigrationsPendingSuggestion matches the migration command suggested by
// rails.
var railsMigrationsPendingSuggestion = regexp.MustCompile(`To resolve this issue, run:\s+(.*?)\n`)

// railsMigrationsPendingMatch reports whether rails has pending migrations.
func railsMigrationsPendingMatch(command *types.Command) bool {
	return strings.Contains(command.Output, "Migrations are pending. To resolve this issue, run:")
}

// railsMigrationsPendingNewCommand runs the pending migrations and retries the
// command.
func railsMigrationsPendingNewCommand(command *types.Command) ([]string, error) {
	match := railsMigrationsPendingSuggestion.FindStringSubmatch(command.Output)
	if match == nil {
		return nil, nil
	}
	return []string{shells.Current().And(match[1], command.Script)}, nil
}

func init() {
	register(types.NewRule("rails_migrations_pending",
		railsMigrationsPendingMatch,
		railsMigrationsPendingNewCommand))
}
