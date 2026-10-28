package rules

import "testing"

// railsMigrationsPendingDevelopmentOutput is the rails output of the Python
// tests for the development environment.
const railsMigrationsPendingDevelopmentOutput = `
Migrations are pending. To resolve this issue, run:

        rails db:migrate RAILS_ENV=development
`

// railsMigrationsPendingTestOutput is the rails output of the Python tests for
// the test environment.
const railsMigrationsPendingTestOutput = `
Migrations are pending. To resolve this issue, run:

        bin/rails db:migrate RAILS_ENV=test
`

func TestRuleRailsMigrationsPending(t *testing.T) {
	resetTestState(t)
	rule := Get("rails_migrations_pending")

	t.Run("match", func(t *testing.T) {
		assertMatch(t, rule, "", railsMigrationsPendingDevelopmentOutput)
		assertMatch(t, rule, "", railsMigrationsPendingTestOutput)
	})

	t.Run("not match", func(t *testing.T) {
		assertNotMatch(t, rule, "Environment data not found in the schema. To resolve this issue, run: \n\n", "")
	})

	t.Run("get new command", func(t *testing.T) {
		assertNewCommands(t, rule, "bin/rspec", railsMigrationsPendingDevelopmentOutput,
			[]string{"rails db:migrate RAILS_ENV=development && bin/rspec"})
		assertNewCommands(t, rule, "bin/rspec", railsMigrationsPendingTestOutput,
			[]string{"bin/rails db:migrate RAILS_ENV=test && bin/rspec"})
	})
}
