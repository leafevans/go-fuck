package rules

import "testing"

// goMissingSumOutput is the output of the Go toolchain for a module that is
// required but missing from go.sum.
func goMissingSumOutput() string {
	return "main.go:3:8: missing go.sum entry for module providing package github.com/pkg/errors (imported by probe); to add:\n" +
		"\tgo get probe\n"
}

func TestRuleGoMissingSum(t *testing.T) {
	resetTestState(t)
	rule := Get("go_missing_sum")
	if rule == nil {
		t.Fatal("go_missing_sum rule is not registered")
	}

	assertMatch(t, rule, "go build ./...", goMissingSumOutput())
	assertMatch(t, rule, "go test ./...", goMissingSumOutput())
	assertNotMatch(t, rule, "cargo build", goMissingSumOutput())
	assertNotMatch(t, rule, "go build ./...", "go: downloading github.com/pkg/errors v0.9.1")

	assertNewCommands(t, rule, "go build ./...", goMissingSumOutput(),
		[]string{"go mod tidy && go build ./..."})
}
