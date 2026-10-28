package rules

import "testing"

func TestRuleNoSuchFile(t *testing.T) {
	resetTestState(t)
	rule := Get("no_such_file")
	if rule == nil {
		t.Fatal("no_such_file rule is not registered")
	}

	t.Run("match", func(t *testing.T) {
		for _, testCase := range []struct {
			script string
			output string
		}{
			{"mv foo bar/foo", "mv: cannot move 'foo' to 'bar/foo': No such file or directory"},
			{"mv foo bar/", "mv: cannot move 'foo' to 'bar/': No such file or directory"},
		} {
			assertMatch(t, rule, testCase.script, testCase.output)
		}
	})

	t.Run("not match", func(t *testing.T) {
		for _, testCase := range []struct {
			script string
			output string
		}{
			{"mv foo bar/", ""},
			{"mv foo bar/foo", "mv: permission denied"},
		} {
			assertNotMatch(t, rule, testCase.script, testCase.output)
		}
	})

	t.Run("new command", func(t *testing.T) {
		for _, testCase := range []struct {
			script string
			output string
			want   []string
		}{
			{"mv foo bar/foo", "mv: cannot move 'foo' to 'bar/foo': No such file or directory",
				[]string{"mkdir -p bar && mv foo bar/foo"}},
			{"mv foo bar/", "mv: cannot move 'foo' to 'bar/': No such file or directory",
				[]string{"mkdir -p bar && mv foo bar/"}},
		} {
			assertNewCommands(t, rule, testCase.script, testCase.output, testCase.want)
		}
	})
}
