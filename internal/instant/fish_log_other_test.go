//go:build !windows

package instant

import "testing"

// fishLog is a compact terminal log written by a logged fish shell. Fish
// redraws the typed command after every keystroke with syntax highlighting and
// cursor moves, so the raw echo line contains the command more than once; the
// echo renderer replays the line before matching.
func fishLog() string {
	return prompt("\x1b[?2004h\r\x1b]0;/repo\x07\rleafevans@host /repo> \x1b[K\r\x1b[47Cgit \r\x1b[51Cpsuh\r\x1b[55C\b\b\b\b\b\b\b\b\x1b[34mgit\x1b[30m \x1b[36mpsuh\x1b[30m\r\x1b[55C\r") + "\n" +
		"\x1b[?2004l\x1b]0;git psuh\x07\rgit: 'psuh' is not a git command. See 'git --help'.\r\n" +
		"\r\nThe most similar command is\r\n\tpush\r\n" +
		prompt("\x1b[?2004h\r\x1b]0;/repo\x07\rleafevans@host /repo> \x1b[K") + "\n"
}

func TestLastOutputFromFishLog(t *testing.T) {
	data, ok := LastOutput(writeLog(t, fishLog()), "git psuh")
	if !ok {
		t.Fatal("LastOutput() did not find the fish command")
	}
	want := "git: 'psuh' is not a git command. See 'git --help'.\n\nThe most similar command is\n\tpush"
	if data != want {
		t.Fatalf("LastOutput() = %q, want %q", data, want)
	}
}

func TestLastOutputFishMiss(t *testing.T) {
	if _, ok := LastOutput(writeLog(t, fishLog()), "git brnch"); ok {
		t.Fatal("LastOutput() found a command that is not in the fish log")
	}
}
