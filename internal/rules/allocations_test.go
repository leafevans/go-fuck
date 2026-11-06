// The race detector instruments every memory access and therefore adds
// allocations that have nothing to do with the rule engine. Disable the
// allocation budgets under -race; the normal test jobs on every platform and
// the coverage job still run them.
//go:build !race

package rules

import (
	"testing"

	"github.com/leafevans/go-fuck/internal/conf"
	"github.com/leafevans/go-fuck/internal/shells"
	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// allocationTestSetup mirrors benchmarkRuleSetup for tests: no user
// configuration, a generic shell and no executables on PATH, so allocation
// measurements do not depend on the machine.
func allocationTestSetup(t *testing.T) {
	t.Helper()
	conf.Reset()
	shells.Set(shells.NewGeneric())
	t.Setenv("PATH", "")
	// On Windows os/exec.LookPath also probes the current directory. Setting
	// this variable disables the probe so that executable lookups fail without
	// touching the file system.
	t.Setenv("NoDefaultCurrentDirectoryInExePath", "1")
	utils.ResetCaches()
	t.Cleanup(func() {
		conf.Reset()
		shells.Set(shells.NewGeneric())
		utils.ResetCaches()
	})
}

// TestRuleMatchingAllocations guards the hot path of go-fuck against memory
// regressions: matching every registered rule against one failed command must
// not start allocating more than a small, stable budget.
//
// Allocations are deterministic across machines and Go versions, unlike
// wall-clock time, so this check can run in the normal test suite on every
// platform. Raise the budget deliberately when a rule legitimately needs more
// allocations; never raise it just to make the test pass.
func TestRuleMatchingAllocations(t *testing.T) {
	allocationTestSetup(t)

	command := types.NewCommand("git push", benchmarkGitPushOutput())
	registered := All()
	const budget = 25

	allocations := testing.AllocsPerRun(20, func() {
		for _, rule := range registered {
			rule.IsMatch(command)
		}
	})
	t.Logf("matching %d rules allocates %.1f times per pass", len(registered), allocations)
	if allocations > budget {
		t.Fatalf("matching %d rules allocates %.1f times per pass, budget is %d", len(registered), allocations, budget)
	}
}

// TestRuleMatchingWithoutOutputAllocations guards the fast path used when the
// command output could not be captured: rules that require the output are
// skipped before their match function runs, so this pass must stay allocation
// free.
func TestRuleMatchingWithoutOutputAllocations(t *testing.T) {
	allocationTestSetup(t)

	command := types.NewCommandWithoutOutput("git push")
	registered := All()

	allocations := testing.AllocsPerRun(50, func() {
		for _, rule := range registered {
			rule.IsMatch(command)
		}
	})
	if allocations != 0 {
		t.Fatalf("matching %d rules without output allocates %.1f times per pass, want 0", len(registered), allocations)
	}
}
