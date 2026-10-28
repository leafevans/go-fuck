package rules

import "testing"

func TestRuleWhois(t *testing.T) {
	resetTestState(t)
	rule := Get("whois")

	t.Run("match", func(t *testing.T) {
		assertMatch(t, rule, "whois https://en.wikipedia.org/wiki/Main_Page", "")
		assertMatch(t, rule, "whois https://en.wikipedia.org/", "")
		assertMatch(t, rule, "whois meta.unix.stackexchange.com", "")
	})

	t.Run("not match", func(t *testing.T) {
		assertNotMatch(t, rule, "whois", "")
	})

	t.Run("get new command", func(t *testing.T) {
		assertNewCommands(t, rule, "whois https://en.wikipedia.org/wiki/Main_Page", "",
			[]string{"whois en.wikipedia.org"})
		assertNewCommands(t, rule, "whois https://en.wikipedia.org/", "",
			[]string{"whois en.wikipedia.org"})
		assertNewCommands(t, rule, "whois meta.unix.stackexchange.com", "",
			[]string{"whois unix.stackexchange.com", "whois stackexchange.com", "whois com"})
	})
}
