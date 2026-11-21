package logs

import (
	"strings"
	"testing"
)

func TestCommandText(t *testing.T) {
	for _, tc := range []struct{ script, want string }{
		{"git status", "git status"},
		{"echo 中文 café", "echo 中文 café"},
		{"echo 'one\ntwo'", `echo 'one\ntwo'`},
		{"echo\tvalue\r\n", `echo\tvalue\r\n`},
		{"echo \x1b[2J\a\b\x00\x7f\u009b", `echo \x1b[2J\a\b\x00\x7f\u009b`},
		{`echo 'literal\n'`, `echo 'literal\n'`},
		{"echo \u202eabc\u2028end\u2029", `echo \u202eabc\u2028end\u2029`},
	} {
		t.Run(tc.script, func(t *testing.T) {
			if got := CommandText(tc.script); got != tc.want {
				t.Fatalf("CommandText(%q) = %q, want %q", tc.script, got, tc.want)
			}
		})
	}
}

func TestCommandPromptsEscapeControls(t *testing.T) {
	buffer := resetLogsState(t)
	script := "echo 'one\ntwo'\x1b[2J"
	ConfirmText(script, true, "reason\ntext", "remove\tfile")
	NewLine()
	ShowCorrectedCommand(script, true, "reason\ntext", "remove\tfile")
	got := buffer.String()
	if strings.Count(got, `echo 'one\ntwo'\x1b[2J`) != 2 {
		t.Fatalf("command displays differ: %q", got)
	}
	if strings.Contains(got, "\x1b[2J") || strings.Count(got, "\n") != 2 {
		t.Fatalf("command controls affected the terminal layout: %q", got)
	}
	if strings.Count(got, `reason\ntext  |  Before running: remove\tfile`) != 2 {
		t.Fatalf("suggestion metadata was not rendered consistently: %q", got)
	}
}
