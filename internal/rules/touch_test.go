package rules

import "testing"

func TestRuleTouch(t *testing.T) {
	resetTestState(t)
	rule := Get("touch")
	if rule == nil {
		t.Fatal("touch rule is not registered")
	}

	output := func(isBSD bool) string {
		if isBSD {
			return "touch: /a/b/c: No such file or directory"
		}
		return "touch: cannot touch '/a/b/c': No such file or directory"
	}

	t.Run("match", func(t *testing.T) {
		assertMatch(t, rule, "touch /a/b/c", output(false))
		assertMatch(t, rule, "touch /a/b/c", output(true))
	})

	t.Run("not match", func(t *testing.T) {
		assertNotMatch(t, rule, "touch /a/b/c", "")
		assertNotMatch(t, rule, "ls /a/b/c", output(false))
	})

	t.Run("new command", func(t *testing.T) {
		for _, isBSD := range []bool{false, true} {
			assertNewCommands(t, rule, "touch /a/b/c", output(isBSD),
				[]string{"mkdir -p /a/b && touch /a/b/c"})
		}
	})
}
