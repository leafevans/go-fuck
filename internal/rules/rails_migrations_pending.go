package rules

import (
	"regexp"
	"strings"

	"github.com/leafevans/go-fuck/v2/internal/shells"
	"github.com/leafevans/go-fuck/v2/internal/types"
)

// railsMigrationsPendingSuggestion matches the migration command suggested by
// rails.
//
// The command sits on its own line after the marker, possibly behind blank
// lines, and the output may end without a final newline. Requiring a non-empty
// line keeps a stray blank line from producing an empty command, which would
// turn the suggestion into " && <command>".
var railsMigrationsPendingSuggestion = regexp.MustCompile(`To resolve this issue, run:[ \t]*\n(?:[ \t]*\n)*[ \t]*([^\n]+)`)

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
	migration := strings.TrimSpace(match[1])
	if migration == "" {
		return nil, nil
	}
	return []string{shells.Current().And(migration, command.Script)}, nil
}

func init() {
	register(types.NewRule("rails_migrations_pending",
		railsMigrationsPendingMatch,
		railsMigrationsPendingNewCommand))
}
