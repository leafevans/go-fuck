package rules

import "testing"

// hostscliNoWebsiteOutput is the hostscli output of the Python tests.
const hostscliNoWebsiteOutput = `hostscli.errors.WebsiteImportError:

No Domain list found for website: a_website_that_does_not_exist

Please raise a Issue here: https://github.com/dhilipsiva/hostscli/issues/new
if you think we should add domains for this website.

type ` + "`hostscli websites`" + ` to see a list of websites that you can block/unblock
`

func TestRuleHostscli(t *testing.T) {
	resetTestState(t)
	rule := Get("hostscli")

	t.Run("match unknown website", func(t *testing.T) {
		assertMatch(t, rule, "hostscli block a_website_that_does_not_exist", hostscliNoWebsiteOutput)
	})

	t.Run("match unknown command", func(t *testing.T) {
		assertMatch(t, rule, "hostscli blok a_website", `Error: No such command "blok"`)
	})

	t.Run("not match", func(t *testing.T) {
		assertNotMatch(t, rule, "hostscli block example.com", "Blocked example.com")
	})

	t.Run("get new command", func(t *testing.T) {
		assertNewCommands(t, rule, "hostscli block a_website_that_does_not_exist", hostscliNoWebsiteOutput,
			[]string{"hostscli websites"})
	})
}
