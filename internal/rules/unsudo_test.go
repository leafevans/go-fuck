package rules

import "testing"

func TestRuleUnsudo(t *testing.T) {
	resetTestState(t)
	rule := Get("unsudo")
	if rule == nil {
		t.Fatal("unsudo rule is not registered")
	}

	t.Run("match", func(t *testing.T) {
		assertMatch(t, rule, "sudo ls", "you cannot perform this operation as root")
	})

	t.Run("not match", func(t *testing.T) {
		assertNotMatch(t, rule, "", "")
		assertNotMatch(t, rule, "sudo ls", "Permission denied")
		assertNotMatch(t, rule, "ls", "you cannot perform this operation as root")
	})

	t.Run("new command", func(t *testing.T) {
		assertNewCommands(t, rule, "sudo ls", "", []string{"ls"})
		assertNewCommands(t, rule, "sudo pacaur -S helloworld", "",
			[]string{"pacaur -S helloworld"})
	})
}
