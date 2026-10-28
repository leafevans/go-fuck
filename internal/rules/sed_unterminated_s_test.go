package rules

import "testing"

func TestRuleSedUnterminatedS(t *testing.T) {
	resetTestState(t)
	rule := Get("sed_unterminated_s")
	if rule == nil {
		t.Fatal("sed_unterminated_s rule is not registered")
	}

	errorOutput := "sed: -e expression #1, char 9: unterminated `s' command"

	t.Run("match", func(t *testing.T) {
		for _, script := range []string{
			"sed -e s/foo/bar",
			"sed -es/foo/bar",
			"sed -e s/foo/bar -e s/baz/quz",
		} {
			assertMatch(t, rule, script, errorOutput)
			assertNotMatch(t, rule, script, "")
		}
	})

	t.Run("new command", func(t *testing.T) {
		for _, testCase := range []struct {
			script string
			want   []string
		}{
			{"sed -e s/foo/bar", []string{"sed -e s/foo/bar/"}},
			{"sed -es/foo/bar", []string{"sed -es/foo/bar/"}},
			{"sed -e 's/\\/foo/bar'", []string{"sed -e 's/\\/foo/bar/'"}},
			{"sed -e s/foo/bar -es/baz/quz",
				[]string{"sed -e s/foo/bar/ -es/baz/quz/"}},
		} {
			assertNewCommands(t, rule, testCase.script, errorOutput, testCase.want)
		}
	})
}
