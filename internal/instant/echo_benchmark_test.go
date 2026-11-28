package instant

import (
	"strings"
	"testing"
)

// BenchmarkEchoIndex isolates the per-row comparison from log rendering.
func BenchmarkEchoIndex(b *testing.B) {
	for _, tc := range []struct {
		name, echo, command string
		want                int
	}{
		{"ExactHit", "$ git status --short", "git status --short", 0},
		{"SpacingHit", "$ git  status\t--short", "git status --short", 0},
		{"UnicodeSpacingHit", "$ git\u2003status\u00a0--short", "git status --short", 0},
		{"ShortMiss", "$ git status --short", "git status --verbose", -1},
		{"LongMiss", "$ echo " + strings.Repeat("x", 256), "echo absent", -1},
	} {
		b.Run(tc.name, func(b *testing.B) {
			call := logCall{prompt: tc.echo}
			parts := strings.Fields(tc.command)
			if got := call.echoIndex(tc.command, parts); got != tc.want {
				b.Fatalf("echoIndex() = %d, want %d", got, tc.want)
			}
			b.ReportAllocs()
			for b.Loop() {
				call.echoIndex(tc.command, parts)
			}
		})
	}
}

// BenchmarkMultilineEchoMiss measures a prompt block with many possible
// starting rows, but no occurrence of the second command line.
func BenchmarkMultilineEchoMiss(b *testing.B) {
	lines := make([]string, 16)
	for i := range lines {
		lines[i] = "prompt echo first line"
	}
	command := "echo first line\necho absent line"
	if got := multilineEchoIndex(lines, command); got != -1 {
		b.Fatalf("multilineEchoIndex() = %d, want -1", got)
	}
	b.ReportAllocs()
	for b.Loop() {
		multilineEchoIndex(lines, command)
	}
}
