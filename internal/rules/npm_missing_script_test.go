package rules

import (
	"testing"

	"github.com/leafevans/go-fuck/internal/specific"
)

// npmMissingScriptOutput returns the npm output that reports a missing script.
func npmMissingScriptOutput(script string) string {
	return `
npm ERR! Linux 4.4.0-31-generic
npm ERR! argv "/opt/node/bin/node" "/opt/node/bin/npm" "run" "dvelop"
npm ERR! node v4.4.7
npm ERR! npm  v2.15.8

npm ERR! missing script: ` + script + `
npm ERR!
npm ERR! If you need help, you may report this error at:
npm ERR!     <https://github.com/npm/npm/issues>

npm ERR! Please include the following file with any support request:
npm ERR!     /home/nvbn/exp/code_view/client_web/npm-debug.log
`
}

// npmMissingScriptStubScripts replaces the npm scripts with the fixture values.
func npmMissingScriptStubScripts(t *testing.T) {
	t.Helper()
	original := specific.NpmGetScripts
	specific.NpmGetScripts = func() []string {
		return []string{"build", "develop", "watch-test"}
	}
	t.Cleanup(func() { specific.NpmGetScripts = original })
}

// npmMissingScriptAssertFirstCommand checks the best suggestion. The original
// test only compares the first corrected command as well.
func npmMissingScriptAssertFirstCommand(t *testing.T, script, output, want string) {
	t.Helper()
	rule := Get("npm_missing_script")
	got, err := rule.GetNewCommand(testCommand(script, output))
	if err != nil {
		t.Fatalf("rule %q returned an error: %v", rule.Name, err)
	}
	if len(got) == 0 {
		t.Fatalf("rule %q returned no corrected commands for %q", rule.Name, script)
	}
	if got[0] != want {
		t.Fatalf("rule %q returned %q, want first command %q", rule.Name, got, want)
	}
}

func TestRuleNpmMissingScript(t *testing.T) {
	resetTestState(t)
	npmMissingScriptStubScripts(t)

	rule := Get("npm_missing_script")
	if rule == nil {
		t.Fatal("npm_missing_script rule is not registered")
	}

	assertMatch(t, rule, "npm ru wach", npmMissingScriptOutput("wach"))
	assertMatch(t, rule, "npm run live-tes", npmMissingScriptOutput("live-tes"))
	assertMatch(t, rule, "npm run-script sahare", npmMissingScriptOutput("sahare"))

	assertNotMatch(t, rule, "npm wach", npmMissingScriptOutput("wach"))
	assertNotMatch(t, rule, "vim live-tes", npmMissingScriptOutput("live-tes"))
	assertNotMatch(t, rule, "npm run-script sahare", "")

	npmMissingScriptAssertFirstCommand(t, "npm ru wach-tests",
		npmMissingScriptOutput("wach-tests"), "npm ru watch-test")
	npmMissingScriptAssertFirstCommand(t, "npm -i run-script dvelop",
		npmMissingScriptOutput("dvelop"), "npm -i run-script develop")
	npmMissingScriptAssertFirstCommand(t, "npm -i run-script buld -X POST",
		npmMissingScriptOutput("buld"), "npm -i run-script build -X POST")
}
