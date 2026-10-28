package rules

import "testing"

// herokuNotCommandSuggestOutput is the heroku output of the Python tests.
const herokuNotCommandSuggestOutput = `
 ▸    log is not a heroku command.
 ▸    Perhaps you meant logs?
 ▸    Run heroku _ to run heroku logs.
 ▸    Run heroku help for a list of available commands.`

func TestRuleHerokuNotCommand(t *testing.T) {
	resetTestState(t)
	rule := Get("heroku_not_command")

	t.Run("match", func(t *testing.T) {
		assertMatch(t, rule, "heroku log", herokuNotCommandSuggestOutput)
	})

	t.Run("not match", func(t *testing.T) {
		assertNotMatch(t, rule, "cat log", herokuNotCommandSuggestOutput)
	})

	t.Run("get new command", func(t *testing.T) {
		assertNewCommands(t, rule, "heroku log", herokuNotCommandSuggestOutput,
			[]string{"heroku logs"})
	})
}
