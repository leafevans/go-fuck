package rules

import "testing"

func TestRuleUnknownCommand(t *testing.T) {
	resetTestState(t)
	rule := Get("unknown_command")
	if rule == nil {
		t.Fatal("unknown_command rule is not registered")
	}

	const hdfsOutput = "ls: Unknown command\nDid you mean -ls?  This command begins with a dash."

	for _, testCase := range []struct {
		script string
		output string
	}{
		{"./bin/hdfs dfs ls", hdfsOutput},
		{"hdfs dfs ls", hdfsOutput},
		{"hdfs dfs ls /foo/bar", hdfsOutput},
	} {
		assertMatch(t, rule, testCase.script, testCase.output)
	}

	for _, testCase := range []struct {
		script string
		output string
	}{
		{"./bin/hdfs dfs -ls", ""},
		{"./bin/hdfs dfs -ls /foo/bar", ""},
		{"hdfs dfs -ls -R /foo/bar", ""},
		{"", ""},
	} {
		assertNotMatch(t, rule, testCase.script, testCase.output)
	}

	for _, testCase := range []struct {
		script string
		output string
		want   []string
	}{
		{"hdfs dfs ls", hdfsOutput, []string{"hdfs dfs -ls"}},
		{"hdfs dfs rm /foo/bar",
			"rm: Unknown command\nDid you mean -rm?  This command begins with a dash.",
			[]string{"hdfs dfs -rm /foo/bar"}},
		{"./bin/hdfs dfs ls -R /foo/bar", hdfsOutput, []string{"./bin/hdfs dfs -ls -R /foo/bar"}},
		{"./bin/hdfs dfs -Dtest=fred ls -R /foo/bar", hdfsOutput,
			[]string{"./bin/hdfs dfs -Dtest=fred -ls -R /foo/bar"}},
	} {
		assertNewCommands(t, rule, testCase.script, testCase.output, testCase.want)
	}
}
