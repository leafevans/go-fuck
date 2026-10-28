package rules

import "testing"

func TestRuleCpCreateDestination(t *testing.T) {
	resetTestState(t)
	rule := Get("cp_create_destination")
	if rule == nil {
		t.Fatal("cp_create_destination rule is not registered")
	}

	t.Run("match", func(t *testing.T) {
		for _, testCase := range []struct {
			script string
			output string
		}{
			{"cp", "cp: directory foo does not exist\n"},
			{"mv", "No such file or directory"},
		} {
			assertMatch(t, rule, testCase.script, testCase.output)
		}
	})

	t.Run("not match", func(t *testing.T) {
		for _, testCase := range []struct {
			script string
			output string
		}{
			{"cp", ""},
			{"mv", ""},
			{"ls", "No such file or directory"},
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
			{"cp foo bar/", "cp: directory foo does not exist\n",
				[]string{"mkdir -p bar/ && cp foo bar/"}},
			{"mv foo bar/", "No such file or directory",
				[]string{"mkdir -p bar/ && mv foo bar/"}},
			{"cp foo bar/baz/", "cp: directory foo does not exist\n",
				[]string{"mkdir -p bar/baz/ && cp foo bar/baz/"}},
		} {
			assertNewCommands(t, rule, testCase.script, testCase.output, testCase.want)
		}
	})
}
