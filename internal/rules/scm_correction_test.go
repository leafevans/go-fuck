package rules

import "testing"

// scmCorrectionSetup installs a fake actual SCM and restores it afterwards.
func scmCorrectionSetup(t *testing.T, scm string) {
	t.Helper()
	original := scmCorrectionActualSCM
	scmCorrectionActualSCM = func() string { return scm }
	t.Cleanup(func() { scmCorrectionActualSCM = original })
}

func TestRuleSCMCorrection(t *testing.T) {
	resetTestState(t)
	rule := Get("scm_correction")
	if rule == nil {
		t.Fatal("scm_correction rule is not registered")
	}

	const gitOutput = "fatal: Not a git repository (or any of the parent directories): .git"
	const hgOutput = "abort: no repository found in '/home/nvbn/exp/thefuck' (.hg not found)!"

	t.Run("match", func(t *testing.T) {
		for _, testCase := range []struct {
			script string
			output string
			actual string
		}{
			{"git log", gitOutput, "hg"},
			{"hg log", hgOutput, "git"},
		} {
			scmCorrectionSetup(t, testCase.actual)
			assertMatch(t, rule, testCase.script, testCase.output)
		}
	})

	t.Run("not match", func(t *testing.T) {
		for _, testCase := range []struct {
			script string
			output string
			actual string
		}{
			{"git log", "", "hg"},
			{"git log", gitOutput, ""},
			{"hg log", hgOutput, ""},
			{"not-scm log", hgOutput, "git"},
		} {
			scmCorrectionSetup(t, testCase.actual)
			assertNotMatch(t, rule, testCase.script, testCase.output)
		}
	})

	t.Run("new command", func(t *testing.T) {
		for _, testCase := range []struct {
			script string
			actual string
			want   string
		}{
			{"git log", "hg", "hg log"},
			{"hg log", "git", "git log"},
		} {
			scmCorrectionSetup(t, testCase.actual)
			assertNewCommands(t, rule, testCase.script, "", []string{testCase.want})
		}
	})
}
