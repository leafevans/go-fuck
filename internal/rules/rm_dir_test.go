package rules

import "testing"

func TestRuleRmDir(t *testing.T) {
	resetTestState(t)
	rule := Get("rm_dir")
	if rule == nil {
		t.Fatal("rm_dir rule is not registered")
	}

	t.Run("match", func(t *testing.T) {
		for _, testCase := range []struct {
			script string
			output string
		}{
			{"rm foo", "rm: foo: is a directory"},
			{"rm foo", "rm: foo: Is a directory"},
			{"hdfs dfs -rm foo", "rm: `foo`: Is a directory"},
			{"./bin/hdfs dfs -rm foo", "rm: `foo`: Is a directory"},
			{"sudo rm foo", "rm: foo: is a directory"},
		} {
			assertMatch(t, rule, testCase.script, testCase.output)
		}
	})

	t.Run("not match", func(t *testing.T) {
		for _, testCase := range []struct {
			script string
			output string
		}{
			{"rm foo", ""},
			{"hdfs dfs -rm foo", ""},
			{"./bin/hdfs dfs -rm foo", ""},
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
			{"rm foo", []string{"rm -rf foo"}},
			{"hdfs dfs -rm foo", []string{"hdfs dfs -rm -r foo"}},
			{"sudo rm foo", []string{"sudo rm -rf foo"}},
		} {
			assertNewCommands(t, rule, testCase.script, "", testCase.want)
		}
	})
}
