package shells

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/leafevans/go-fuck/internal/conf"
)

// BenchmarkReadHistoryFile measures parsing a large shell history file.
//
// The history based rules are matched for every command, so this parser is on
// the hot path even when the failed command has nothing to do with history.
func BenchmarkReadHistoryFile(b *testing.B) {
	path := filepath.Join(b.TempDir(), "history")
	lines := make([]string, 20000)
	for i := range lines {
		lines[i] = "git status --short --branch"
	}
	if err := os.WriteFile(path, []byte(strings.Join(lines, string(rune(charNewline)))+string(rune(charNewline))), 0o644); err != nil {
		b.Fatalf("write history: %v", err)
	}

	conf.Reset()
	b.Cleanup(conf.Reset)
	conf.Current().HistoryLimit = 0

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		readHistoryFile(path, func(line string) string { return line })
	}
}
