package utils

import (
	"fmt"
	"testing"
)

// buildExecutableNames returns a deterministic list of names that resembles a
// large PATH listing.
func buildExecutableNames(count int) []string {
	const alphabet = "abcdefghijklmnopqrstuvwxyz"
	names := make([]string, 0, count)
	for i := range count {
		length := 3 + i%12
		name := make([]byte, 0, length)
		for j := range length {
			name = append(name, alphabet[(i*7+j*3)%len(alphabet)])
		}
		names = append(names, string(name))
	}
	return names
}

// BenchmarkGetCloseMatchesLarge measures the misspelled command suggestion
// search over a PATH sized candidate list, which is the hot path of the
// no_command and history rules.
func BenchmarkGetCloseMatchesLarge(b *testing.B) {
	possibilities := buildExecutableNames(3000)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetCloseMatches("gti", possibilities, 3, 0.6)
	}
}

// BenchmarkGetCloseMatchesLowCutoff measures the same search with the low
// cutoff used by command specific rules such as git_not_command.
func BenchmarkGetCloseMatchesLowCutoff(b *testing.B) {
	possibilities := buildExecutableNames(3000)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetCloseMatches("gti", possibilities, 3, 0.1)
	}
}

// BenchmarkSequenceRatio measures the raw similarity helper.
func BenchmarkSequenceRatio(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		SequenceRatio("git push --set-upstream origin main", "git branch")
	}
}

var _ = fmt.Sprintf
