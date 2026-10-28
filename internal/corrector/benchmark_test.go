package corrector

import (
	"testing"

	"github.com/leafevans/go-fuck/internal/conf"
	"github.com/leafevans/go-fuck/internal/shells"
	"github.com/leafevans/go-fuck/internal/types"
	"github.com/leafevans/go-fuck/internal/utils"
)

// benchmarkCorrectorSetup resets the process wide state so that benchmarks do
// not depend on the user configuration, on the shell or on the installed
// tools.
//
// PATH is empty so that executable lookups fail before touching the file
// system: the benchmarks must not depend on which executables are installed,
// must not start external processes and must not use the network. This keeps
// the measurements deterministic on CI machines and on developer machines.
func benchmarkCorrectorSetup(b *testing.B) {
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

// BenchmarkGetRules benchmarks collecting and ordering the enabled rules with
// the default settings.
func BenchmarkGetRules(b *testing.B) {
	benchmarkCorrectorSetup(b)

	b.ReportAllocs()
	b.ResetTimer()
	for iteration := 0; iteration < b.N; iteration++ {
		GetRules()
	}
}

// BenchmarkGetCorrectedCommands benchmarks the complete correction hot path:
// selecting the enabled rules, matching them against the failed command and
// organizing the corrections.
func BenchmarkGetCorrectedCommands(b *testing.B) {
	benchmarkCorrectorSetup(b)
	command := types.NewCommand("git push", benchmarkGitPushOutput())

	b.ReportAllocs()
	b.ResetTimer()
	for iteration := 0; iteration < b.N; iteration++ {
		GetCorrectedCommands(command)
	}
}
