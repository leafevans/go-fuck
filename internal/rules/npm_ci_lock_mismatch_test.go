package rules

import "testing"

// npmCILockMismatchOutput is the npm 7 and newer output for a lock file that
// is out of sync with package.json. npm 7 and 8 prefix the lines with
// "npm ERR!" and npm 9 and newer with "npm error".
func npmCILockMismatchOutput(prefix string) string {
	return prefix + " code EUSAGE\n" +
		prefix + "\n" +
		prefix + " `npm ci` can only install packages when your package.json and package-lock.json or npm-shrinkwrap.json are in sync. Please update your lock file with `npm install` before continuing.\n" +
		prefix + "\n" +
		prefix + " Missing: lodash@4.17.21 from lock file\n" +
		prefix + "\n" +
		prefix + " Clean install a project\n"
}

func TestRuleNpmCILockMismatch(t *testing.T) {
	resetTestState(t)
	rule := Get("npm_ci_lock_mismatch")
	if rule == nil {
		t.Fatal("npm_ci_lock_mismatch rule is not registered")
	}

	assertMatch(t, rule, "npm ci", npmCILockMismatchOutput("npm error"))
	assertMatch(t, rule, "npm ci --omit=dev", npmCILockMismatchOutput("npm ERR!"))
	assertNotMatch(t, rule, "pnpm ci", npmCILockMismatchOutput("npm error"))
	assertNotMatch(t, rule, "npm ci", "")

	assertNewCommands(t, rule, "npm ci", npmCILockMismatchOutput("npm error"),
		[]string{"npm install && npm ci"})
}
