package rules

import "testing"

// tsuruLoginErrorMessages are the error messages of the Python tests.
var tsuruLoginErrorMessages = []string{
	"Error: you're not authenticated or your session has expired.",
	"You're not authenticated or your session has expired. Please use \"login\" command for authentication.",
}

func TestRuleTsuruLogin(t *testing.T) {
	resetTestState(t)
	rule := Get("tsuru_login")

	t.Run("match", func(t *testing.T) {
		assertMatch(t, rule, "tsuru app-shell", tsuruLoginErrorMessages[0])
		assertMatch(t, rule, "tsuru app-log -f", tsuruLoginErrorMessages[1])
	})

	t.Run("not match", func(t *testing.T) {
		assertNotMatch(t, rule, "tsuru", "")
		assertNotMatch(t, rule, "tsuru app-restart", "Error: unauthorized")
		assertNotMatch(t, rule, "tsuru app-log -f", "Error: unparseable data")
	})

	t.Run("get new command", func(t *testing.T) {
		assertNewCommands(t, rule, "tsuru app-shell", tsuruLoginErrorMessages[0],
			[]string{"tsuru login && tsuru app-shell"})
		assertNewCommands(t, rule, "tsuru app-log -f", tsuruLoginErrorMessages[1],
			[]string{"tsuru login && tsuru app-log -f"})
	})
}
