package rules

import (
	"testing"

	"github.com/leafevans/go-fuck/internal/conf"
	"github.com/leafevans/go-fuck/internal/shells"
	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// benchmarkRuleSetup resets the process wide state so that benchmarks do not
// depend on the user configuration, on the shell or on the installed tools.
//
// PATH is empty so that executable lookups fail before touching the file
// system: the benchmarks must not depend on which executables are installed,
// must not start external processes and must not use the network. This keeps
// the measurements deterministic on CI machines and on developer machines.
func benchmarkRuleSetup(b *testing.B) {
	b.Helper()
	conf.Reset()
	shells.Set(shells.NewGeneric())
	b.Setenv("PATH", "")
	// On Windows os/exec.LookPath also probes the current directory. Setting
	// this variable disables the probe, so executable lookups in the benchmark
	// fail without touching the file system. It is ignored on other platforms.
	b.Setenv("NoDefaultCurrentDirectoryInExePath", "1")
	utils.ResetCaches()
	b.Cleanup(func() {
		conf.Reset()
		shells.Set(shells.NewGeneric())
		utils.ResetCaches()
	})
}

// benchmarkGitPushOutput is a realistic output of a failed "git push".
func benchmarkGitPushOutput() string {
	return "fatal: The current branch master has no upstream branch.\n" +
		"To push the current branch and set the remote as upstream, use\n\n" +
		"    git push --set-upstream origin master\n\n"
}

// BenchmarkRuleMatchingAllRules benchmarks the hottest path of go-fuck: every
// registered rule is tried against the failed command. The default settings
// enable all rules. The chosen command only matches rules whose match function
// is pure, so the benchmark stays deterministic.
func BenchmarkRuleMatchingAllRules(b *testing.B) {
	benchmarkRuleSetup(b)
	command := types.NewCommand("git push", benchmarkGitPushOutput())
	registered := All()

	b.ReportAllocs()
	b.ResetTimer()
	for iteration := 0; iteration < b.N; iteration++ {
		for _, rule := range registered {
			rule.IsMatch(command)
		}
	}
}

// BenchmarkRuleMatchingWithoutOutput benchmarks the fast path used when the
// command output could not be captured: rules that require the output are
// skipped before their match function runs.
func BenchmarkRuleMatchingWithoutOutput(b *testing.B) {
	benchmarkRuleSetup(b)
	command := types.NewCommandWithoutOutput("git push")
	registered := All()

	b.ReportAllocs()
	b.ResetTimer()
	for iteration := 0; iteration < b.N; iteration++ {
		for _, rule := range registered {
			rule.IsMatch(command)
		}
	}
}
