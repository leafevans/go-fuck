package rules

import "testing"

// goInstallVersionOutput is the output of go install outside of a module.
func goInstallVersionOutput() string {
	return "go: 'go install' requires a version when current directory is not in a module\n" +
		"\tTry 'go install example.com/foo/bar@latest' to install the latest version\n"
}

func TestRuleGoInstallVersion(t *testing.T) {
	resetTestState(t)
	rule := Get("go_install_version")
	if rule == nil {
		t.Fatal("go_install_version rule is not registered")
	}

	assertMatch(t, rule, "go install example.com/foo/bar", goInstallVersionOutput())
	assertNotMatch(t, rule, "cargo install foo", goInstallVersionOutput())
	assertNotMatch(t, rule, "go install example.com/foo/bar@latest", "go: downloading")

	assertNewCommands(t, rule, "go install example.com/foo/bar", goInstallVersionOutput(),
		[]string{"go install example.com/foo/bar@latest"})
}
