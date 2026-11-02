package rules

import (
	"testing"

	"github.com/leafevans/go-fuck/internal/specific"
)

// makeUnknownTargetOutput returns the output of GNU make for a missing target.
func makeUnknownTargetOutput(target string) string {
	return "make: *** No rule to make target '" + target + "'.  Stop.\n"
}

// stubMakeTargets replaces the Makefile targets with fixture values.
func stubMakeTargets(t *testing.T) {
	t.Helper()
	original := specific.MakeTargets
	specific.MakeTargets = func() []string {
		return []string{"build", "clean", "lint", "test"}
	}
	t.Cleanup(func() { specific.MakeTargets = original })
}

func TestRuleMakeUnknownTarget(t *testing.T) {
	resetTestState(t)
	stubMakeTargets(t)
	rule := Get("make_unknown_target")
	if rule == nil {
		t.Fatal("make_unknown_target rule is not registered")
	}

	assertMatch(t, rule, "make buidl", makeUnknownTargetOutput("buidl"))
	assertMatch(t, rule, "make -j4 buidl", makeUnknownTargetOutput("buidl"))
	assertNotMatch(t, rule, "cmake buidl", makeUnknownTargetOutput("buidl"))
	assertNotMatch(t, rule, "make build", "make: Nothing to be done for 'build'.\n")

	got, err := rule.GetNewCommand(testCommand("make buidl", makeUnknownTargetOutput("buidl")))
	if err != nil {
		t.Fatalf("rule %q returned an error: %v", rule.Name, err)
	}
	if len(got) == 0 || got[0] != "make build" {
		t.Fatalf("rule %q returned %q, want first command %q", rule.Name, got, "make build")
	}
}

func TestRuleMakeUnknownTargetNeededBy(t *testing.T) {
	resetTestState(t)
	stubMakeTargets(t)
	rule := Get("make_unknown_target")

	// A dependency that is not part of the command line cannot be corrected by
	// replacing an argument, so the rule must not match it.
	output := "make: *** No rule to make target 'buidl', needed by 'all'.  Stop.\n"
	assertNotMatch(t, rule, "make all", output)

	// When the missing target is on the command line it is still corrected.
	got, err := rule.GetNewCommand(testCommand("make all buidl", output))
	if err != nil {
		t.Fatalf("rule %q returned an error: %v", rule.Name, err)
	}
	if len(got) == 0 || got[0] != "make all build" {
		t.Fatalf("rule %q returned %q, want first command %q", rule.Name, got, "make all build")
	}
}

func TestRuleMakeUnknownTargetWithoutTargets(t *testing.T) {
	resetTestState(t)
	original := specific.MakeTargets
	specific.MakeTargets = func() []string { return nil }
	t.Cleanup(func() { specific.MakeTargets = original })

	rule := Get("make_unknown_target")
	assertNotMatch(t, rule, "make buidl", makeUnknownTargetOutput("buidl"))
}
