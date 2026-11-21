package specific

import (
	"strings"
	"testing"

	"github.com/leafevans/go-fuck/v2/internal/shells"
)

func BenchmarkGitOptionSpan(b *testing.B) {
	for _, tc := range []struct{ name, script string }{
		{"plain", "git status -short"},
		{"global", "git -C 'project dir' -c color.ui=no status -short"},
		{"paths", "git status -short -- " + strings.Repeat("'long path/file.txt' ", 64)},
	} {
		b.Run(tc.name, func(b *testing.B) {
			shell := shells.NewBash()
			b.ReportAllocs()
			for b.Loop() {
				if _, _, ok := GitOptionSpan(tc.script, "-short", shell); !ok {
					b.Fatal("option not found")
				}
			}
		})
	}
}
