package rules

import "testing"

// npmCIMissingLockfileOutput is the npm output of npm ci without a lock file.
// npm 7 and 8 prefix the lines with "npm ERR!" and npm 9 and newer with
// "npm error".
func npmCIMissingLockfileOutput(prefix string) string {
	return prefix + " code EUSAGE\n" +
		prefix + "\n" +
		prefix + " The `npm ci` command can only install with an existing package-lock.json or\n" +
		prefix + " npm-shrinkwrap.json with lockfileVersion >= 1. Run an install with npm@5 or\n" +
		prefix + " later to generate a package-lock.json file, then try again.\n" +
		prefix + "\n" +
		prefix + " Clean install a project\n"
}

func TestRuleNpmCIMissingLockfile(t *testing.T) {
	resetTestState(t)
	rule := Get("npm_ci_missing_lockfile")
	if rule == nil {
		t.Fatal("npm_ci_missing_lockfile rule is not registered")
	}

	assertMatch(t, rule, "npm ci", npmCIMissingLockfileOutput("npm error"))
	assertMatch(t, rule, "npm ci --omit=dev", npmCIMissingLockfileOutput("npm ERR!"))
	assertNotMatch(t, rule, "pnpm ci", npmCIMissingLockfileOutput("npm error"))
	assertNotMatch(t, rule, "npm ci", "")
	assertNotMatch(t, rule, "npm install", "\n`npm ci` can only install packages when your package.json and package-lock.json or npm-shrinkwrap.json are in sync.\n")

	assertNewCommands(t, rule, "npm ci", npmCIMissingLockfileOutput("npm error"),
		[]string{"npm install && npm ci"})
}
