//go:build windows

package instant

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/leafevans/go-fuck/v4/internal/consts"
)

// BenchmarkConPTY replays a synthetic terminal log rather than timing a live
// shell. Every case consumes the same bytes, with file I/O included only in
// the LastOutput cases.
func BenchmarkConPTY(b *testing.B) {
	const (
		calls      = 512
		diagnostic = "error: 无法加载 module; check the import path"
		advice     = "hint: verify the project configuration and try again"
	)
	var fixture strings.Builder
	fixture.WriteString(conPTYSizeRecord + " 120 30\n")
	for i := range calls {
		// Erase and redraw a stale echo, then use an absolute cursor
		// column to overwrite the start of the advice row.
		fmt.Fprintf(&fixture, "%sPS> build mod\r\x1b[2K%sPS> build module-%04d\r\n",
			consts.UserCommandMark, consts.UserCommandMark, i)
		fixture.WriteString("\x1b[31m" + diagnostic + "\x1b[0m\r\nwait\x1b[1G" + advice + "\r\n")
	}
	fixture.WriteString(consts.UserCommandMark + "PS> ")
	data := fixture.String()

	b.Run("Render", func(b *testing.B) {
		// Check that a redraw and enough scrolling to reach the last call
		// preserve the expected prompt and diagnostic lines.
		lines := logLines(data)
		wantPrompt := fmt.Sprintf("%sPS> build module-%04d", consts.UserCommandMark, calls-1)
		index := slices.Index(lines, wantPrompt)
		if index < 0 || index+2 >= len(lines) || lines[index+1] != diagnostic || lines[index+2] != advice {
			b.Fatal("rendered log does not contain the final command and its diagnostics")
		}
		b.ReportAllocs()
		b.SetBytes(int64(len(data)))
		for b.Loop() {
			logLines(data)
		}
	})

	for _, tc := range []struct {
		name, script, want string
		hit                bool
	}{
		{"LastOutputHit", fmt.Sprintf("build module-%04d", calls-1), diagnostic + "\n" + advice, true},
		{"LastOutputMiss", "build module-absent", "", false},
	} {
		b.Run(tc.name, func(b *testing.B) {
			path := filepath.Join(b.TempDir(), "conpty.log")
			if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
				b.Fatal(err)
			}
			if output, ok := LastOutput(path, tc.script); output != tc.want || ok != tc.hit {
				b.Fatalf("LastOutput(%q) = (%q, %t), want (%q, %t)", tc.script, output, ok, tc.want, tc.hit)
			}
			b.ReportAllocs()
			b.SetBytes(int64(len(data)))
			for b.Loop() {
				LastOutput(path, tc.script)
			}
		})
	}
}
