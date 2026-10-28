package rules

import "testing"

func TestRulePhpS(t *testing.T) {
	resetTestState(t)
	rule := Get("php_s")

	t.Run("match", func(t *testing.T) {
		assertMatch(t, rule, "php -s localhost:8000", "")
		assertMatch(t, rule, "php -t pub -s 0.0.0.0:8080", "")
	})

	t.Run("not match", func(t *testing.T) {
		assertNotMatch(t, rule, "php -S localhost:8000", "")
		assertNotMatch(t, rule, "vim php -s", "")
		assertNotMatch(t, rule, "php -s", "")
	})

	t.Run("get new command", func(t *testing.T) {
		assertNewCommands(t, rule, "php -s localhost:8000", "", []string{"php -S localhost:8000"})
		assertNewCommands(t, rule, "php -t pub -s 0.0.0.0:8080", "", []string{"php -t pub -S 0.0.0.0:8080"})
	})
}
