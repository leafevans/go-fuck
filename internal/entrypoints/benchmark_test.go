package entrypoints

import "testing"

// BenchmarkMarkedBashRawCommand benchmarks restoring a multi-line command from
// the output of "fc -ln -2".
func BenchmarkMarkedBashRawCommand(b *testing.B) {
	payload := "\t if true; then\necho hi\nfi\n\t fuck"
	b.ReportAllocs()
	b.ResetTimer()
	for iteration := 0; iteration < b.N; iteration++ {
		bashMarkedRawCommand(payload)
	}
}

// BenchmarkMarkedZshRawCommand benchmarks restoring a command from the zsh
// history array.
func BenchmarkMarkedZshRawCommand(b *testing.B) {
	payload := "if true; then\necho hi\nfi\n"
	b.ReportAllocs()
	b.ResetTimer()
	for iteration := 0; iteration < b.N; iteration++ {
		zshMarkedRawCommand(payload)
	}
}
