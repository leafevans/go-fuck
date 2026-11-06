package instant

import "testing"

// ANSI escape sequences of the realistic prompt fixtures. They are stored in
// constants so that the fixtures below stay readable.
const (
	promptGreen = "\x1b[0;32m"
	promptCyan  = "\x1b[1;36m"
	promptBlue  = "\x1b[1;34m"
	promptReset = "\x1b[0m"
	promptTitle = "\x1b]0;user@host: ~/project\x07"
)

// TestLastOutputPromptMatrix checks that LastOutput finds the typed command and
// returns its exact output for the prompt styles that the POSIX shells render
// into the instant mode log.
//
// The fixtures contain the bytes that the pseudo terminal writes, so they mix
// the instant mode prompt marker (a carriage return resets the column after the
// invisible marker), ANSI colors, an OSC window title, CRLF and LF line
// endings and backspace redraws. Multi line prompt themes start the prompt on
// a fresh line, which keeps the marker on its own log line.
func TestLastOutputPromptMatrix(t *testing.T) {
	command := "git brnch"
	tests := []struct {
		name   string
		log    string
		script string
		want   string
		found  bool
	}{
		{
			name: "bash bira multi line prompt",
			log: prompt("") + "\r\n" +
				promptGreen + "╭─ user@host ~/project" + promptReset + "\r\n" +
				promptGreen + "╰─$ " + promptReset + command + "\r\n" +
				"git: 'brnch' is not a git command. See 'git --help'.\r\n" +
				prompt("") + "\r\n" +
				promptGreen + "╭─ user@host ~/project" + promptReset + "\r\n" +
				promptGreen + "╰─$ " + promptReset,
			script: command,
			want:   "git: 'brnch' is not a git command. See 'git --help'.",
			found:  true,
		},
		{
			name: "bash starship single line prompt",
			log: prompt(promptGreen+"➜"+promptReset+" "+promptCyan+"project"+promptReset+" "+
				promptBlue+"git:(main)"+promptReset+" ") + command + "\r\n" +
				"git: 'brnch' is not a git command. See 'git --help'.\r\n" +
				prompt(promptGreen+"➜"+promptReset+" "+promptCyan+"project"+promptReset+" "),
			script: command,
			want:   "git: 'brnch' is not a git command. See 'git --help'.",
			found:  true,
		},
		{
			name: "zsh color prompt",
			log: prompt("") + "\r\n" +
				promptGreen + "╭─" + promptReset + promptCyan + " user@host " + promptReset +
				promptBlue + "~/project" + promptReset + "\r\n" +
				promptGreen + "╰─$ " + promptReset + "git status\r\n" +
				"fatal: not a git repository (or any of the parent directories): .git\r\n" +
				prompt("") + "\r\n" +
				promptGreen + "╭─" + promptReset + promptCyan + " user@host " + promptReset +
				promptBlue + "~/project" + promptReset + "\r\n" +
				promptGreen + "╰─$ " + promptReset,
			script: "git status",
			want:   "fatal: not a git repository (or any of the parent directories): .git",
			found:  true,
		},
		{
			name: "osc window title before the marker",
			log: promptTitle + prompt(promptGreen+"$ git brnch"+promptReset) + "\r\n" +
				promptTitle + "git: 'brnch' is not a git command." + "\r\n" +
				promptTitle + prompt("$ "),
			script: command,
			want:   "git: 'brnch' is not a git command.",
			found:  true,
		},
		{
			name: "crlf line endings",
			log: prompt("$ echo hello") + "\r\n" +
				"hello\r\n" +
				prompt("$ "),
			script: "echo hello",
			want:   "hello",
			found:  true,
		},
		{
			name: "lf line endings",
			log: prompt("$ uname -s") + "\n" +
				"Linux\n" +
				prompt("$ "),
			script: "uname -s",
			want:   "Linux",
			found:  true,
		},
		{
			name: "backspace redraw artifacts",
			log: prompt("$ ") + "g\bgit brnch" + "\r\n" +
				"git: 'brnch' is not a git command." + "\r\n" +
				prompt("$ "),
			script: command,
			want:   "git: 'brnch' is not a git command.",
			found:  true,
		},
		{
			name: "command absent from the log",
			log: prompt("$ ls") + "\r\n" +
				"file one\r\n" +
				prompt("$ "),
			script: command,
			want:   "",
			found:  false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, ok := LastOutput(writeLog(t, test.log), test.script)
			if ok != test.found {
				t.Fatalf("LastOutput() found = %t, want %t (output %q)", ok, test.found, got)
			}
			if got != test.want {
				t.Fatalf("LastOutput() = %q, want %q", got, test.want)
			}
		})
	}
}
