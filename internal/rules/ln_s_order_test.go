package rules

import "testing"

func TestRuleLnSOrder(t *testing.T) {
	resetTestState(t)
	rule := Get("ln_s_order")
	if rule == nil {
		t.Fatal("ln_s_order rule is not registered")
	}

	originalPathExists := lnSOrderPathExists
	t.Cleanup(func() { lnSOrderPathExists = originalPathExists })

	output := func(filename string) string {
		return "ln: failed to create symbolic link '" + filename + "': File exists"
	}

	t.Run("not match", func(t *testing.T) {
		for _, testCase := range []struct {
			script string
			output string
			exists bool
		}{
			{"ln dest source", output("source"), true},
			{"ls -s dest source", output("source"), true},
			{"ln -s dest source", "", true},
			{"ln -s dest source", output("source"), false},
		} {
			lnSOrderPathExists = func(string) bool { return testCase.exists }
			assertNotMatch(t, rule, testCase.script, testCase.output)
		}
	})

	t.Run("match and new command", func(t *testing.T) {
		lnSOrderPathExists = func(string) bool { return true }
		for _, testCase := range []struct {
			script string
			want   []string
		}{
			{"ln -s dest source", []string{"ln -s source dest"}},
			{"ln dest -s source", []string{"ln -s source dest"}},
			{"ln dest source -s", []string{"ln source -s dest"}},
			{"sudo ln -s dest source", []string{"sudo ln -s source dest"}},
		} {
			assertMatch(t, rule, testCase.script, output("source"))
			assertNewCommands(t, rule, testCase.script, output("source"), testCase.want)
		}
	})
}
