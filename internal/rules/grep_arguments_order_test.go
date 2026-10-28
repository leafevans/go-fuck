package rules

import (
	"strings"
	"testing"
)

func TestRuleGrepArgumentsOrder(t *testing.T) {
	resetTestState(t)
	rule := Get("grep_arguments_order")
	if rule == nil {
		t.Fatal("grep_arguments_order rule is not registered")
	}

	originalPathExists := grepArgumentsOrderPathExists
	t.Cleanup(func() { grepArgumentsOrderPathExists = originalPathExists })
	// Python test: monkeypatch.setattr('os.path.isfile', lambda x: not
	// x.startswith('-'))
	grepArgumentsOrderPathExists = func(path string) bool {
		return !strings.HasPrefix(path, "-")
	}

	output := func(file string) string {
		return "grep: " + file + ": No such file or directory"
	}

	t.Run("match", func(t *testing.T) {
		for _, testCase := range []struct {
			script string
			file   string
		}{
			{"grep test.py test", "test"},
			{"grep -lir . test", "test"},
			{"egrep test.py test", "test"},
			{"egrep -lir . test", "test"},
		} {
			assertMatch(t, rule, testCase.script, output(testCase.file))
		}
	})

	t.Run("not match", func(t *testing.T) {
		for _, testCase := range []struct {
			script string
			output string
		}{
			{"cat test.py", output("test")},
			{"grep test test.py", ""},
			{"grep -lir test .", ""},
			{"egrep test test.py", ""},
			{"egrep -lir test .", ""},
		} {
			assertNotMatch(t, rule, testCase.script, testCase.output)
		}
	})

	t.Run("new command", func(t *testing.T) {
		for _, testCase := range []struct {
			script string
			file   string
			want   []string
		}{
			{"grep test.py test", "test", []string{"grep test test.py"}},
			{"grep -lir . test", "test", []string{"grep -lir test ."}},
			{"grep . test -lir", "test", []string{"grep test -lir ."}},
			{"egrep test.py test", "test", []string{"egrep test test.py"}},
			{"egrep -lir . test", "test", []string{"egrep -lir test ."}},
			{"egrep . test -lir", "test", []string{"egrep test -lir ."}},
		} {
			assertNewCommands(t, rule, testCase.script, output(testCase.file),
				testCase.want)
		}
	})
}
