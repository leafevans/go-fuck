package rules

import "testing"

func TestRuleQuotationMarks(t *testing.T) {
	rule := Get("quotation_marks")
	if rule == nil {
		t.Fatal("quotation_marks rule is not registered")
	}

	for _, script := range []string{
		"git commit -m 'My Message\"",
		"echo 'hello\"",
	} {
		assertMatch(t, rule, script, "")
	}

	for _, script := range []string{
		"git commit -m 'My Message'",
		"git commit -m \"My Message\"",
		"echo hello",
	} {
		assertNotMatch(t, rule, script, "")
	}

	assertNewCommands(t, rule, "git commit -m 'My Message\"", "",
		[]string{"git commit -m \"My Message\""})
}
