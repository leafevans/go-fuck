package rules

import "testing"

func TestRuleCpOmittingDirectory(t *testing.T) {
	resetTestState(t)
	rule := Get("cp_omitting_directory")
	if rule == nil {
		t.Fatal("cp_omitting_directory rule is not registered")
	}

	t.Run("match", func(t *testing.T) {
		for _, testCase := range []struct {
			script string
			output string
		}{
			{"cp dir", "cp: dor: is a directory"},
			{"cp dir", "cp: omitting directory 'dir'"},
			{"sudo cp dir", "cp: omitting directory 'dir'"},
		} {
			assertMatch(t, rule, testCase.script, testCase.output)
		}
	})

	t.Run("not match", func(t *testing.T) {
		for _, testCase := range []struct {
			script string
			output string
		}{
			{"some dir", "cp: dor: is a directory"},
			{"some dir", "cp: omitting directory 'dir'"},
			{"cp dir", ""},
		} {
			assertNotMatch(t, rule, testCase.script, testCase.output)
		}
	})

	t.Run("new command", func(t *testing.T) {
		assertNewCommands(t, rule, "cp dir", "", []string{"cp -a dir"})
		assertNewCommands(t, rule, "sudo cp dir", "", []string{"sudo cp -a dir"})
	})
}
