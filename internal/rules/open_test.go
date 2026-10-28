package rules

import "testing"

func TestRuleOpen(t *testing.T) {
	resetTestState(t)
	rule := Get("open")
	if rule == nil {
		t.Fatal("open rule is not registered")
	}

	output := func(script string) string {
		for index, char := range script {
			if char == ' ' {
				return "The file " + script[index+1:] + " does not exist.\n"
			}
		}
		return ""
	}

	t.Run("is arg url", func(t *testing.T) {
		for _, script := range []string{
			"open foo.com", "open foo.edu", "open foo.info", "open foo.io",
			"open foo.ly", "open foo.me", "open foo.net", "open foo.org",
			"open foo.se", "open www.foo.ru",
		} {
			if !openIsArgURL(testCommand(script, "")) {
				t.Errorf("openIsArgURL should report true for %q", script)
			}
		}
		for _, script := range []string{"open foo", "open bar.txt", "open egg.doc"} {
			if openIsArgURL(testCommand(script, "")) {
				t.Errorf("openIsArgURL should report false for %q", script)
			}
		}
	})

	t.Run("match", func(t *testing.T) {
		for _, script := range []string{
			"open foo.com", "xdg-open foo.com", "gnome-open foo.com",
			"kde-open foo.com", "open nonest",
		} {
			assertMatch(t, rule, script, output(script))
		}
	})

	t.Run("not match", func(t *testing.T) {
		assertNotMatch(t, rule, "open nonest", "")
		assertNotMatch(t, rule, "ls nonest", "The file nonest does not exist.\n")
	})

	t.Run("new command", func(t *testing.T) {
		for _, testCase := range []struct {
			script string
			want   []string
		}{
			{"open foo.io", []string{"open http://foo.io"}},
			{"xdg-open foo.io", []string{"xdg-open http://foo.io"}},
			{"gnome-open foo.io", []string{"gnome-open http://foo.io"}},
			{"kde-open foo.io", []string{"kde-open http://foo.io"}},
			{"open nonest", []string{
				"touch nonest && open nonest",
				"mkdir nonest && open nonest",
			}},
		} {
			assertNewCommands(t, rule, testCase.script, output(testCase.script), testCase.want)
		}
	})
}
