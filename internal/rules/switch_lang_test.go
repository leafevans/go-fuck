package rules

import "testing"

func TestRuleSwitchLang(t *testing.T) {
	resetTestState(t)
	t.Setenv("TF_ALIAS", "fuck")

	rule := Get("switch_lang")
	if rule == nil {
		t.Fatal("switch_lang rule is not registered")
	}

	t.Run("match", func(t *testing.T) {
		for _, testCase := range []struct {
			script string
			output string
		}{
			{"\u0444\u0437\u0435-\u043f\u0443\u0435", "command not found: \u0444\u0437\u0435-\u043f\u0443\u0435"},
			{"\u03bb\u03c3", "command not found: \u03bb\u03c3"},
			{"\u05e9\u05e4\u05d0-\u05e2\u05e7\u05d0", "command not found: \u05e9\u05e4\u05d0-\u05e2\u05e7\u05d0"},
			{"\u05da\u05d3", "command not found: \u05da\u05d3"},
			{"\ub140\uc560 \u3163\u3134", "command not found: \ub140\uc560 \u3163\u3134"},
		} {
			assertMatch(t, rule, testCase.script, testCase.output)
		}
	})

	t.Run("not match", func(t *testing.T) {
		for _, testCase := range []struct {
			script string
			output string
		}{
			{"pat-get", "command not found: pat-get"},
			{"ls", "command not found: ls"},
			{"\u0430\u0433\u0441\u043b", "command not found: \u0430\u0433\u0441\u043b"},
			{"\u0444\u0437\u0435-\u043f\u0443\u0435", "some info"},
			{"\u05e9\u05e4\u05d0-\u05e2\u05e7\u05d0", "some info"},
			{"\ub140\uc560 \u3163\u3134", "some info"},
		} {
			assertNotMatch(t, rule, testCase.script, testCase.output)
		}
	})

	t.Run("new command", func(t *testing.T) {
		for _, testCase := range []struct {
			script string
			output string
			want   string
		}{
			{"\u0444\u0437\u0435-\u043f\u0443\u0435 \u0448\u0442\u044b\u0435\u0444\u0434\u0434 \u043c\u0448\u044c", "", "apt-get install vim"},
			{"\u03bb\u03c3 -\u03bb\u03b1", "", "ls -la"},
			{"\u05e9\u05e4\u05d0-\u05e2\u05e7\u05d0 \u05df\u05de\u05d3\u05d0\u05e9\u05da\u05da \u05d4\u05df\u05e6", "", "apt-get install vim"},
			{"\u05da\u05d3 -\u05da\u05e9", "", "ls -la"},
			{"\uba67-\u314e\u3137\u3145 \u3151\u315c\u3134\u3145\ubbf8\u3163 \ud344\u3161", "", "apt-get install vim"},
			{"\u3163\u3134 -\u3163\u3141", "", "ls -la"},
			{"\u3154\u3141\u3145\ucd24", "", "patchk"},
		} {
			assertNewCommands(t, rule, testCase.script, testCase.output, []string{testCase.want})
		}
	})
}
