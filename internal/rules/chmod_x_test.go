package rules

import "testing"

func TestRuleChmodX(t *testing.T) {
	resetTestState(t)
	rule := Get("chmod_x")
	if rule == nil {
		t.Fatal("chmod_x rule is not registered")
	}

	originalExists := chmodXPathExists
	originalExecutable := chmodXIsExecutable
	t.Cleanup(func() {
		chmodXPathExists = originalExists
		chmodXIsExecutable = originalExecutable
	})
	chmodXPathExists = func(string) bool { return true }
	chmodXIsExecutable = func(string) bool { return false }

	t.Run("match", func(t *testing.T) {
		for _, testCase := range []struct {
			script string
			output string
		}{
			{"./gradlew build", "gradlew: Permission denied"},
			{"./install.sh --help", "install.sh: permission denied"},
		} {
			assertMatch(t, rule, testCase.script, testCase.output)
		}
	})

	t.Run("not match", func(t *testing.T) {
		for _, testCase := range []struct {
			script   string
			output   string
			exists   bool
			callable bool
		}{
			{"./gradlew build", "gradlew: Permission denied", true, true},
			{"./gradlew build", "gradlew: Permission denied", false, false},
			{"./gradlew build", "gradlew: error", true, false},
			{"gradlew build", "gradlew: Permission denied", true, false},
		} {
			chmodXPathExists = func(string) bool { return testCase.exists }
			chmodXIsExecutable = func(string) bool { return testCase.callable }
			assertNotMatch(t, rule, testCase.script, testCase.output)
		}
		chmodXPathExists = func(string) bool { return true }
		chmodXIsExecutable = func(string) bool { return false }
	})

	t.Run("new command", func(t *testing.T) {
		for _, testCase := range []struct {
			script string
			want   []string
		}{
			{"./gradlew build", []string{"chmod +x gradlew && ./gradlew build"}},
			{"./install.sh --help", []string{"chmod +x install.sh && ./install.sh --help"}},
		} {
			assertNewCommands(t, rule, testCase.script, "", testCase.want)
		}
	})
}
