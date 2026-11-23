package specific

import (
	"testing"

	"github.com/leafevans/go-fuck/v3/internal/shells"
)

func TestGitArgumentSpanPreservesShellWords(t *testing.T) {
	for _, tc := range []struct {
		name, script, argument string
		shell                  shells.Shell
		want                   string
	}{
		{"bash path", `git -C 'my repo' log master -- 'master'`, "master", shells.NewBash(), "master"},
		{"bash quoted", `git log 'master' -- 'master'`, "master", shells.NewBash(), "'master'"},
		{"powershell backtick", "git -C my` repo log 'master' -- master", "master", shells.NewPowershell(), "'master'"},
		{"powershell quote", `git -C 'it''s repo' log master -- master`, "master", shells.NewPowershell(), "master"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			start, end, ok := GitArgumentSpan(tc.script, tc.argument, tc.shell, false)
			if !ok || tc.script[start:end] != tc.want {
				t.Fatalf("span = %d:%d, %v; want %q", start, end, ok, tc.want)
			}
		})
	}
	for _, script := range []string{"git log master master", "git log -- master", "git log HEAD && echo master", "git log 'HEAD'; echo master", "git log master 'unfinished", "git log # master"} {
		if _, _, ok := GitArgumentSpan(script, "master", shells.NewBash(), false); ok {
			t.Errorf("unexpected match for %q", script)
		}
	}
}

func TestGitArgumentSpanAcrossShells(t *testing.T) {
	for _, tc := range []struct {
		name  string
		shell shells.Shell
	}{
		{"bash", shells.NewBash()},
		{"generic", shells.NewGeneric()},
		{"zsh", shells.NewZsh()},
		{"fish", shells.NewFish()},
		{"tcsh", shells.NewTcsh()},
		{"powershell", shells.NewPowershell()},
	} {
		t.Run(tc.name, func(t *testing.T) {
			argument := "master's branch"
			cwd := `repo\it's dir`
			quoted := tc.shell.Quote(argument)
			script := "git -C " + tc.shell.Quote(cwd) + " log " + quoted + " -- " + quoted
			start, end, ok := GitArgumentSpan(script, argument, tc.shell, false)
			if !ok || script[start:end] != quoted {
				t.Fatalf("span = %d:%d, %v in %q; want %q", start, end, ok, script, quoted)
			}
		})
	}
}
