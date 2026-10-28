package rules

import "testing"

func TestRuleCatDir(t *testing.T) {
	resetTestState(t)
	rule := Get("cat_dir")
	if rule == nil {
		t.Fatal("cat_dir rule is not registered")
	}

	originalIsDir := catDirIsDir
	t.Cleanup(func() { catDirIsDir = originalIsDir })

	t.Run("match", func(t *testing.T) {
		catDirIsDir = func(string) bool { return true }
		for _, testCase := range []struct {
			script string
			output string
		}{
			{"cat foo", "cat: foo: Is a directory\n"},
			{"cat /foo/bar/", "cat: /foo/bar/: Is a directory\n"},
			{"cat cat/", "cat: cat/: Is a directory\n"},
		} {
			assertMatch(t, rule, testCase.script, testCase.output)
		}
	})

	t.Run("not match", func(t *testing.T) {
		catDirIsDir = func(string) bool { return false }
		for _, testCase := range []struct {
			script string
			output string
		}{
			{"cat foo", "foo bar baz"},
			{"cat foo bar", "foo bar baz"},
			{"notcat foo bar", "some output"},
		} {
			assertNotMatch(t, rule, testCase.script, testCase.output)
		}
	})

	t.Run("new command", func(t *testing.T) {
		catDirIsDir = func(string) bool { return true }
		for _, testCase := range []struct {
			script string
			output string
			want   []string
		}{
			{"cat foo", "cat: foo: Is a directory\n", []string{"ls foo"}},
			{"cat /foo/bar/", "cat: /foo/bar/: Is a directory\n", []string{"ls /foo/bar/"}},
			{"cat cat", "cat: cat: Is a directory\n", []string{"ls cat"}},
		} {
			assertNewCommands(t, rule, testCase.script, testCase.output, testCase.want)
		}
	})
}
