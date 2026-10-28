package rules

import "testing"

// goMissingModuleOutput is the output of the Go toolchain for an import that
// no required module provides.
func goMissingModuleOutput() string {
	return "main.go:3:8: no required module provides package github.com/definitely/not-a-real-package-xyz; to add it:\n" +
		"\tgo get github.com/definitely/not-a-real-package-xyz\n"
}

func TestRuleGoMissingModule(t *testing.T) {
	resetTestState(t)
	rule := Get("go_missing_module")
	if rule == nil {
		t.Fatal("go_missing_module rule is not registered")
	}

	assertMatch(t, rule, "go build ./...", goMissingModuleOutput())
	assertNotMatch(t, rule, "cargo build", goMissingModuleOutput())
	assertNotMatch(t, rule, "go build ./...", "go: cannot find main module, but found .git/config in C:\\repo")

	assertNewCommands(t, rule, "go build ./...", goMissingModuleOutput(),
		[]string{"go get github.com/definitely/not-a-real-package-xyz && go build ./..."})
}
