package specific

import (
	"testing"

	"github.com/leafevans/go-fuck/v3/internal/shells"
	"github.com/leafevans/go-fuck/v3/internal/types"
)

func TestExpandGitAliasesPreservesArguments(t *testing.T) {
	shell := shells.NewGeneric()
	cases := []struct {
		name, script, trace, want string
	}{
		{
			name:   "repeated alias in arguments",
			script: "git st -- st 'st/file.txt'",
			trace:  "trace: alias expansion: st => status\n",
			want:   "git status -- st 'st/file.txt'",
		},
		{
			name:   "literal dollar signs in expansion",
			script: "git cm -- file.txt",
			trace:  "trace: alias expansion: cm => commit -m '$1 ${USER}'\n",
			want:   "git commit -m '$1 ${USER}' -- file.txt",
		},
		{
			name:   "punctuation in alias",
			script: "git ci! ci!",
			trace:  "trace: alias expansion: ci! => commit\n",
			want:   "git commit ci!",
		},
		{
			name:   "quoted executable and alias",
			script: "'/usr/local/bin/git'\t'st' -- 'st'",
			trace:  "trace: alias expansion: st => status\n",
			want:   "'/usr/local/bin/git'\tstatus -- 'st'",
		},
		{
			name:   "unrelated trace",
			script: "git log -- st",
			trace:  "trace: alias expansion: st => status\n",
			want:   "git log -- st",
		},
		{
			name:   "global option values are preserved",
			script: "git -C 'st folder' -c alias.st=status --no-pager st -- st",
			trace:  "trace: alias expansion: st => status\n",
			want:   "git -C 'st folder' -c alias.st=status --no-pager status -- st",
		},
		{
			name:   "alias used as global option value",
			script: "git --work-tree st log -- st",
			trace:  "trace: alias expansion: st => status\n",
			want:   "git --work-tree st log -- st",
		},
		{
			name:   "no subcommand",
			script: "git -C st",
			trace:  "trace: alias expansion: st => status\n",
			want:   "git -C st",
		},
		{
			name:   "argument separator before alias",
			script: "git -- st",
			trace:  "trace: alias expansion: st => status\n",
			want:   "git -- st",
		},
		{
			name:   "invalid invocation quoting",
			script: "git 'st",
			trace:  "trace: alias expansion: st => status\n",
			want:   "git 'st",
		},
		{
			name:   "shell alias expansion",
			script: "git st",
			trace:  "trace: alias expansion: st => !git status\n",
			want:   "git st",
		},
		{
			name:   "malformed trace",
			script: "git st",
			trace:  "trace: alias expansion: malformed\n",
			want:   "git st",
		},
		{
			name:   "invalid expansion",
			script: "git cm",
			trace:  "trace: alias expansion: cm => commit 'unterminated\n",
			want:   "git cm",
		},
		{
			name:   "empty alias",
			script: "git status",
			trace:  "trace: alias expansion:  => status\n",
			want:   "git status",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			command := types.NewCommand(tc.script, tc.trace)
			got := expandGitAliasesForShell(command, shell)
			if got.Script != tc.want {
				t.Fatalf("script = %q, want %q", got.Script, tc.want)
			}
			if got.Output != tc.trace || !got.HasOutput() {
				t.Fatal("alias expansion changed the captured output")
			}
			if command.Script != tc.script {
				t.Fatal("alias expansion mutated the original command")
			}
		})
	}
}

func TestExpandGitAliasesQuotesForCurrentShell(t *testing.T) {
	for _, shell := range []shells.Shell{shells.NewGeneric(), shells.NewPowershell(), shells.NewFish()} {
		t.Run(shell.Name(), func(t *testing.T) {
			command := types.NewCommand("git cm file.txt", "trace: alias expansion: cm => commit -m 'hello $USER'\n")
			want := "git " + shell.Quote("commit") + " " + shell.Quote("-m") + " " + shell.Quote("hello $USER") + " file.txt"
			if got := expandGitAliasesForShell(command, shell).Script; got != want {
				t.Fatalf("script = %q, want %q", got, want)
			}
		})
	}
}
