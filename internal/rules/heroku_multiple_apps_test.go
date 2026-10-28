package rules

import "testing"

// herokuMultipleAppsSuggestOutput is the heroku output of the Python tests.
const herokuMultipleAppsSuggestOutput = `
 ▸    Multiple apps in git remotes
 ▸    Usage: --remote heroku-dev
 ▸    or: --app myapp-dev
 ▸    Your local git repository has more than 1 app referenced in git remotes.
 ▸    Because of this, we can't determine which app you want to run this command against.
 ▸    Specify the app you want with --app or --remote.
 ▸    Heroku remotes in repo:
 ▸    myapp (heroku)
 ▸    myapp-dev (heroku-dev)
 ▸
 ▸    https://devcenter.heroku.com/articles/multiple-environments
`

// herokuMultipleAppsNotMatchOutput is a heroku output without app suggestions.
const herokuMultipleAppsNotMatchOutput = `
=== HEROKU_POSTGRESQL_TEAL_URL, DATABASE_URL
Plan:                  Hobby-basic
Status:                Available
Connections:           20/20
PG Version:            9.6.4
Created:               2017-01-01 00:00 UTC
Data Size:             99.9 MB
Tables:                99
Rows:                  12345/10000000 (In compliance)
Fork/Follow:           Unsupported
Rollback:              Unsupported
Continuous Protection: Off
Add-on:                postgresql-round-12345
`

func TestRuleHerokuMultipleApps(t *testing.T) {
	resetTestState(t)
	rule := Get("heroku_multiple_apps")

	t.Run("match", func(t *testing.T) {
		assertMatch(t, rule, "heroku pg", herokuMultipleAppsSuggestOutput)
	})

	t.Run("not match", func(t *testing.T) {
		assertNotMatch(t, rule, "heroku pg", herokuMultipleAppsNotMatchOutput)
	})

	t.Run("get new command", func(t *testing.T) {
		assertNewCommands(t, rule, "heroku pg", herokuMultipleAppsSuggestOutput,
			[]string{"heroku pg --app myapp", "heroku pg --app myapp-dev"})
	})
}
