package specific

import (
	"testing"
	"unicode/utf8"

	"github.com/leafevans/go-fuck/v4/internal/shells"
)

// FuzzGitQuotedArgument checks that every supported shell round-trips an
// argument without losing its byte boundary or changing unrelated text.
func FuzzGitQuotedArgument(f *testing.F) {
	for index, argument := range []string{"master", "repo dir", "it's", `back\slash`, "", "line\nbreak", "careful!"} {
		f.Add(argument, uint8(index%6))
	}
	f.Fuzz(func(t *testing.T, argument string, shellIndex uint8) {
		if !utf8.ValidString(argument) || argument == "--" {
			// The shell quote implementations operate on runes; an invalid byte
			// sequence cannot be represented losslessly as shell text. "--" is
			// the pathspec separator regardless of how it is quoted.
			return
		}
		shell := []shells.Shell{
			shells.NewBash(), shells.NewFish(), shells.NewPowershell(),
			shells.NewZsh(), shells.NewTcsh(), shells.NewGeneric(),
		}[shellIndex%6]
		quoted := shell.Quote(argument)
		script := "git -C " + shell.Quote("project dir") + " log " + quoted + " -- " + quoted
		start, end, ok := GitArgumentSpan(script, argument, shell, false)
		if !ok {
			t.Fatalf("GitArgumentSpan() did not find %q in %q", argument, script)
		}
		if start < 0 || end > len(script) || start >= end || script[start:end] != quoted {
			t.Fatalf("invalid span %d:%d for %q in %q", start, end, argument, script)
		}
	})
}
