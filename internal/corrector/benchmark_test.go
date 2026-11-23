package corrector

import (
	"strconv"
	"testing"

	"github.com/leafevans/go-fuck/v3/internal/conf"
	"github.com/leafevans/go-fuck/v3/internal/shells"
	"github.com/leafevans/go-fuck/v3/internal/types"
	"github.com/leafevans/go-fuck/v3/internal/utils"
)

func BenchmarkOrganizeCommands(b *testing.B) {
	for _, size := range []int{10, 100, 1000} {
		b.Run(strconv.Itoa(size), func(b *testing.B) {
			conf.Reset()
			b.Cleanup(conf.Reset)
			commands := make([]*types.CorrectedCommand, size)
			for i := range commands {
				commands[i] = &types.CorrectedCommand{Script: "command " + strconv.Itoa(i), Priority: size - i}
			}
			b.ReportAllocs()
			b.ResetTimer()
			for b.Loop() {
				OrganizeCommands(commands)
			}
		})
	}
}

func BenchmarkOrganizeCommandsSideEffects(b *testing.B) {
	for _, size := range []int{100, 1000} {
		b.Run(strconv.Itoa(size), func(b *testing.B) {
			conf.Reset()
			b.Cleanup(conf.Reset)
			commands := make([]*types.CorrectedCommand, 0, 4*size)
			for _, effect := range []types.SideEffectFunc{nil, organizeEffectOne, organizeEffectTwo, organizeEffectOne} {
				for i := range size {
					commands = append(commands, &types.CorrectedCommand{
						Script: "command " + strconv.Itoa(i), SideEffect: effect, Priority: 100,
					})
				}
			}
			b.ReportAllocs()
			for b.Loop() {
				OrganizeCommands(commands)
			}
		})
	}
}

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
	for b.Loop() {
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
	for b.Loop() {
		GetCorrectedCommands(command)
	}
}
