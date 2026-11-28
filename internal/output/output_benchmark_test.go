package output

import (
	"testing"

	"github.com/leafevans/go-fuck/v4/internal/conf"
)

var benchmarkEnvironment []string

func BenchmarkMergedEnv(b *testing.B) {
	b.Run("defaults", func(b *testing.B) {
		overrides := conf.Current().Env
		b.ReportAllocs()
		for b.Loop() {
			benchmarkEnvironment = mergedEnv(overrides)
		}
	})
	b.Run("no_overrides", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			benchmarkEnvironment = mergedEnv(nil)
		}
	})
}
