package rules

import "testing"

func TestRuleMkdirP(t *testing.T) {
	resetTestState(t)
	rule := Get("mkdir_p")
	if rule == nil {
		t.Fatal("mkdir_p rule is not registered")
	}

	hdfsOutput := "mkdir: `foo/bar/baz\\': No such file or directory"

	t.Run("match", func(t *testing.T) {
		for _, testCase := range []struct {
			script string
			output string
		}{
			{"mkdir foo/bar/baz", "mkdir: foo/bar: No such file or directory"},
			{"./bin/hdfs dfs -mkdir foo/bar/baz", hdfsOutput},
			{"hdfs dfs -mkdir foo/bar/baz", hdfsOutput},
			{"sudo mkdir foo/bar/baz", "mkdir: foo/bar: No such file or directory"},
		} {
			assertMatch(t, rule, testCase.script, testCase.output)
		}
	})

	t.Run("not match", func(t *testing.T) {
		for _, testCase := range []struct {
			script string
			output string
		}{
			{"mkdir foo/bar/baz", ""},
			{"mkdir foo/bar/baz", "foo bar baz"},
			{"hdfs dfs -mkdir foo/bar/baz", ""},
			{"./bin/hdfs dfs -mkdir foo/bar/baz", ""},
			{"", ""},
		} {
			assertNotMatch(t, rule, testCase.script, testCase.output)
		}
	})

	t.Run("new command", func(t *testing.T) {
		for _, testCase := range []struct {
			script string
			want   []string
		}{
			{"mkdir foo/bar/baz", []string{"mkdir -p foo/bar/baz"}},
			{"hdfs dfs -mkdir foo/bar/baz", []string{"hdfs dfs -mkdir -p foo/bar/baz"}},
			{"./bin/hdfs dfs -mkdir foo/bar/baz", []string{"./bin/hdfs dfs -mkdir -p foo/bar/baz"}},
			{"sudo mkdir foo/bar/baz", []string{"sudo mkdir -p foo/bar/baz"}},
		} {
			assertNewCommands(t, rule, testCase.script, "", testCase.want)
		}
	})
}
