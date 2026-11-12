package instant

import (
	"strings"
	"testing"

	"github.com/leafevans/go-fuck/internal/consts"
)

// fuzzHasStrippable reports whether line contains a byte that stripOSC, ansiCSI
// or ansiControl remove. sanitize removes the prompt marker before the escape
// and control sequences, so a marker that is split by such a sequence can be
// joined back together and stay in the output (for example: five marker runes,
// an OSC sequence, five marker runes). A real terminal log never splits the
// prompt marker, so the marker check is only applied to lines without
// strippable bytes.
func fuzzHasStrippable(line string) bool {
	for index := 0; index < len(line); index++ {
		b := line[index]
		if b == 0x1b || (b < 0x20 && b != '\t' && b != '\n') || b == 0x7f {
			return true
		}
	}
	return false
}

// FuzzSanitize checks that sanitize never panics and that it strips escape
// sequences, control characters and the prompt marker from arbitrary terminal
// log lines.
func FuzzSanitize(f *testing.F) {
	mark := consts.UserCommandMark
	half := len(mark) / 2
	seeds := []string{
		"",
		"plain text",
		"\x1b[31mfile one\x1b[0m",
		"\x1b]0;window title\x07file one",
		"progress 1\rprogress 100",
		"tab\tkept and newline\nkept",
		"\x1b[2K\x1b[1;1Hcursor",
		"\x07bell",
		"\x00nul\x08backspace\x1funit\x7fdelete",
		"\x1b",
		"\x1b]unterminated",
		"café 你好",
		mark,
		mark + "$ git brnch",
		mark[:half] + "\x00" + mark[half:],
		mark[:half] + "\x1b]0;title\x07" + mark[half:],
	}
	for _, seed := range seeds {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, line string) {
		got := sanitize(line)
		if len(got) > len(line) {
			t.Fatalf("sanitize(%q) grew the line from %d to %d bytes", line, len(line), len(got))
		}
		for index := 0; index < len(got); index++ {
			b := got[index]
			if b == 0x1b || (b < 0x20 && b != '\t' && b != '\n') || b == 0x7f {
				t.Fatalf("sanitize(%q) = %q leaves control byte 0x%02x", line, got, b)
			}
		}
		if !fuzzHasStrippable(line) && strings.Contains(got, mark) {
			t.Fatalf("sanitize(%q) = %q leaves the prompt marker", line, got)
		}
		// Even when stripping escape sequences joined marker fragments, a
		// second pass removes the marker for good.
		if second := sanitize(got); strings.Contains(second, mark) {
			t.Fatalf("sanitize(sanitize(%q)) = %q leaves the prompt marker", line, second)
		}
	})
}

// FuzzStripOSC checks that stripOSC never panics and only removes bytes.
func FuzzStripOSC(f *testing.F) {
	seeds := []string{
		"",
		"plain text",
		"\x1b]0;window title\x07file one",
		"\x1b]0;title\x07\x1b]1;other\x07content",
		"\x1b]unterminated",
		"\x1b",
		"\x1b\x1b]0;title\x07]",
		"café 你好",
	}
	for _, seed := range seeds {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, line string) {
		got := stripOSC(line)
		if len(got) > len(line) {
			t.Fatalf("stripOSC(%q) grew the string from %d to %d bytes", line, len(line), len(got))
		}
		if !fuzzIsSubsequence(got, line) {
			t.Fatalf("stripOSC(%q) = %q is not a subsequence of the input", line, got)
		}
		if !strings.ContainsRune(line, 0x1b) && got != line {
			t.Fatalf("stripOSC(%q) = %q modified a line without escape characters", line, got)
		}
	})
}

// fuzzIsSubsequence reports whether part can be obtained from whole by
// deleting bytes only.
func fuzzIsSubsequence(part, whole string) bool {
	index := 0
	for i := 0; i < len(whole) && index < len(part); i++ {
		if whole[i] == part[index] {
			index++
		}
	}
	return index == len(part)
}

// FuzzLastOutput checks that arbitrary terminal logs and scripts do not panic
// the log reader and that the rendered output cannot leak the prompt marker.
func FuzzLastOutput(f *testing.F) {
	mark := consts.UserCommandMark
	f.Add(mark+"$ git brnch\r\ngit: 'brnch' is not a git command.\r\n"+mark+"$ ", "git brnch")
	f.Add(mark+"$ ls\nfile one\n"+mark+"$ ", "ls")
	f.Add("", "")
	f.Add(string([]byte{0xff, 0xfe, 0x1b, '[', '3', '1', 'm'}), string([]byte{0xff}))
	f.Fuzz(func(t *testing.T, logData, script string) {
		calls := groupCalls(logLines(logData))
		for _, call := range calls {
			index := call.echoIndex(script, strings.Fields(script))
			if index < 0 {
				continue
			}
			output := call.output(index)
			if strings.Contains(output, mark) {
				t.Fatalf("rendered output %q leaked the prompt marker", output)
			}
		}
	})
}
