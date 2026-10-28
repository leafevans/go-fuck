package rules

import "testing"

func TestRuleLnNoHardLink(t *testing.T) {
	resetTestState(t)
	rule := Get("ln_no_hard_link")
	if rule == nil {
		t.Fatal("ln_no_hard_link rule is not registered")
	}

	const errorMessage = "hard link not allowed for directory"

	t.Run("match", func(t *testing.T) {
		for _, testCase := range []struct {
			script string
			output string
		}{
			{"ln barDir barLink", "ln: ‘barDir’: " + errorMessage},
			{"sudo ln a b", "ln: ‘a’: " + errorMessage},
			{"sudo ln -nbi a b", "ln: ‘a’: " + errorMessage},
		} {
			assertMatch(t, rule, testCase.script, testCase.output)
		}
	})

	t.Run("not match", func(t *testing.T) {
		for _, testCase := range []struct {
			script string
			output string
		}{
			{"", ""},
			{"ln a b", "... hard link"},
			{"sudo ln a b", "... hard link"},
			{"a b", errorMessage},
		} {
			assertNotMatch(t, rule, testCase.script, testCase.output)
		}
	})

	t.Run("new command", func(t *testing.T) {
		for _, testCase := range []struct {
			script string
			want   []string
		}{
			{"ln barDir barLink", []string{"ln -s barDir barLink"}},
			{"sudo ln barDir barLink", []string{"sudo ln -s barDir barLink"}},
			{"sudo ln -nbi a b", []string{"sudo ln -s -nbi a b"}},
			{"ln -nbi a b && ls", []string{"ln -s -nbi a b && ls"}},
			{"ln a ln", []string{"ln -s a ln"}},
			{"sudo ln a ln", []string{"sudo ln -s a ln"}},
		} {
			assertNewCommands(t, rule, testCase.script, "", testCase.want)
		}
	})
}
