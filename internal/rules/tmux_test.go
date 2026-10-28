package rules

import "testing"

// tmuxAmbiguousOutput is the tmux error message of the Python tests.
const tmuxAmbiguousOutput = "ambiguous command: list, could be: " +
	"list-buffers, list-clients, list-commands, list-keys, " +
	"list-panes, list-sessions, list-windows"

func TestRuleTmux(t *testing.T) {
	resetTestState(t)
	rule := Get("tmux")

	t.Run("match", func(t *testing.T) {
		assertMatch(t, rule, "tmux list", tmuxAmbiguousOutput)
	})

	t.Run("not match", func(t *testing.T) {
		assertNotMatch(t, rule, "tmux list", "")
		assertNotMatch(t, rule, "tmux list-sessions", "list-sessions")
	})

	t.Run("get new command", func(t *testing.T) {
		assertNewCommands(t, rule, "tmux list", tmuxAmbiguousOutput,
			[]string{"tmux list-keys", "tmux list-panes", "tmux list-windows"})
	})
}
