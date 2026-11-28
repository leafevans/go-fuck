package instant

import (
	"bytes"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/leafevans/go-fuck/v4/internal/consts"
)

func TestStripOSCTerminators(t *testing.T) {
	for _, tc := range []struct {
		name, input, want string
	}{
		{"BEL", "before\x1b]0;title\x07after", "beforeafter"},
		{"ST", "before\x1b]0;title\x1b\\after", "beforeafter"},
		{"hyperlink", "\x1b]8;;https://example.com\x1b\\file.txt\x1b]8;;\x1b\\: missing", "file.txt: missing"},
		{"mixed", "\x1b]0;title\x07one\x1b]1;icon\x1b\\two", "onetwo"},
		{"embedded escape", "\x1b]0;title\x1b[31m\x1b\\visible", "visible"},
		{"unterminated", "before\x1b]0;title", "before"},
		{"truncated escape", "before\x1b]0;title\x1b", "before"},
		{"plain escape", "before\x1b[31mafter", "before\x1b[31mafter"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := stripOSC(tc.input); got != tc.want {
				t.Fatalf("stripOSC() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestLastOutputPreservesHyperlinkText(t *testing.T) {
	log := prompt("$ cat missing.txt") + "\r\n" +
		"cat: \x1b]8;;file:///tmp/missing.txt\x1b\\missing.txt\x1b]8;;\x1b\\: No such file or directory\r\n" + prompt("$ ")
	want := "cat: missing.txt: No such file or directory"
	if got, ok := LastOutput(writeLog(t, log), "cat missing.txt"); !ok || got != want {
		t.Fatalf("LastOutput() = %q, %t, want %q, true", got, ok, want)
	}
}

func TestLastOutputRejectsSimilarInvocations(t *testing.T) {
	for _, tc := range []struct {
		echo, script string
	}{
		{"git status --short", "git status"},
		{"git status", "git stat"},
		{"git push origin", "git origin push"},
		{"megagit status", "git status"},
		{"git-status", "git status"},
	} {
		t.Run(tc.echo, func(t *testing.T) {
			log := prompt("$ "+tc.echo) + "\r\ncommand output\r\n" + prompt("$ ")
			if got, ok := LastOutput(writeLog(t, log), tc.script); ok {
				t.Fatalf("LastOutput(%q) matched %q and returned %q", tc.script, tc.echo, got)
			}
		})
	}
}

func TestEchoIndexDoesNotSearchDistantOutput(t *testing.T) {
	call := logCall{prompt: prompt("$ different-command")}
	call.lines = append(strings.Split(strings.Repeat("output\n", 20), "\n"), "git brnch")
	if got := call.echoIndex("git brnch", []string{"git", "brnch"}); got != -1 {
		t.Fatalf("echoIndex() = %d, want distant output ignored", got)
	}
}

func TestCommandSuffixBoundaries(t *testing.T) {
	for _, tc := range []struct {
		line string
		want bool
	}{
		{"git status", true},
		{"PS>git status", true},
		{"❯git status", true},
		{"$ git status  ", false},
		{"notgit status", false},
		{"你好git status", false},
		{"1git status", false},
		{"/usr/bin/git status", false},
	} {
		if got := hasCommandSuffix(tc.line, "git status"); got != tc.want {
			t.Errorf("hasCommandSuffix(%q) = %t, want %t", tc.line, got, tc.want)
		}
	}
}

func TestSanitizeEchoPlainAndRedrawn(t *testing.T) {
	for _, tc := range []struct{ input, want string }{
		{consts.UserCommandMark + "PS> echo 世界  ", "PS> echo 世界  "},
		{"PS> echo stale\rPS> echo fresh\x1b[K", "PS> echo fresh"},
		{"PS> echo \xffok", "PS> echo ok"},
		{"PS> echo \ufffd", "PS> echo \ufffd"},
	} {
		if got := sanitizeEcho(tc.input); got != tc.want {
			t.Errorf("sanitizeEcho(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestSanitizeEchoBackspaceEdits(t *testing.T) {
	for _, tc := range []struct{ input, want string }{
		{"$ g\bgit brnch", "$ git brnch"},
		{"$ git brnx\bch", "$ git brnch"},
		{"$ 你\bgit brnch", "$ git brnch"},
		{"\b\bgit brnch", "git brnch"},
	} {
		if got := sanitizeEcho(tc.input); got != tc.want {
			t.Errorf("sanitizeEcho(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestRingWriterLargeWritesKeepNewestBytes(t *testing.T) {
	for _, size := range []int{consts.LogSizeInBytes - 1, consts.LogSizeInBytes, consts.LogSizeInBytes + 1, 2 * consts.LogSizeInBytes} {
		t.Run(strconv.Itoa(size), func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "ring.log")
			file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o600)
			if err != nil {
				t.Fatal(err)
			}
			defer file.Close()
			writer := NewRingWriter(file)
			if _, err := writer.Write([]byte("old data")); err != nil {
				t.Fatal(err)
			}
			data := bytes.Repeat([]byte("a"), size)
			copy(data[size-4:], "TAIL")
			if n, err := writer.Write(data); err != nil || n != size {
				t.Fatalf("Write(%d bytes) = %d, %v", size, n, err)
			}
			got, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			want := append([]byte("old data"), data...)
			want = want[max(0, len(want)-consts.LogSizeInBytes):]
			if !bytes.Equal(got, want) {
				t.Fatalf("log contains %d bytes, want the newest %d bytes", len(got), len(want))
			}
			if _, err := writer.Write([]byte("next")); err != nil {
				t.Fatal(err)
			}
			got, err = os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if len(got) > consts.LogSizeInBytes || !bytes.HasSuffix(got, []byte("TAILnext")) {
				t.Fatalf("subsequent write lost data or exceeded the limit: %d bytes", len(got))
			}
		})
	}
}

func TestRingWriterReportsWriteFailure(t *testing.T) {
	path := writeLog(t, "keep this log")
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	if n, err := NewRingWriter(file).Write([]byte("new")); n != 0 || err == nil {
		t.Fatalf("Write() = %d, %v, want 0 and a write error", n, err)
	}
	if got, err := os.ReadFile(path); err != nil || string(got) != "keep this log" {
		t.Fatalf("failed write changed the log: %q, %v", got, err)
	}
}

func TestLastOutputUnreadableLog(t *testing.T) {
	if got, ok := LastOutput(t.TempDir(), "git status"); ok || got != "" {
		t.Fatalf("LastOutput(directory) = %q, %t, want no output", got, ok)
	}
}

func TestFieldSuffixMatchesFields(t *testing.T) {
	for _, line := range []string{"", "PS> git push", "PS> git\tpush  ", "git\u2003push", "gitpush", "push", "你 git push", "git\xffpush", "\n git \r\n push\u00a0"} {
		for _, parts := range [][]string{nil, {"git", "push"}, {"push"}, {"PS>", "git", "push"}, {"gitpush"}, {"git\xffpush"}} {
			fields := strings.Fields(line)
			want := len(parts) > 0 && len(fields) >= len(parts)
			if want {
				for i, p := range parts {
					if fields[len(fields)-len(parts)+i] != p {
						want = false
						break
					}
				}
			}
			if got := hasFieldSuffix(line, parts); got != want {
				t.Errorf("line=%q parts=%q got=%t want=%t", line, parts, got, want)
			}
		}
	}
}

func TestLastOutputMatchesUnicodeFieldSpacing(t *testing.T) {
	log := prompt("$ git\u2003status\u00a0--short") + "\r\n M example.txt\r\n" + prompt("$ ")
	if got, ok := LastOutput(writeLog(t, log), "git status --short"); !ok || got != " M example.txt" {
		t.Fatalf("LastOutput() = %q, %t; want %q, true", got, ok, " M example.txt")
	}
}

func TestMultilineEchoPromptLimit(t *testing.T) {
	lines := make([]string, 17)
	lines[0], lines[15] = "alpha", "beta"
	if got := multilineEchoIndex(lines, "alpha\nbeta"); got != 15 {
		t.Fatalf("last row inside prompt: echo index %d, want 15", got)
	}
	lines[15], lines[16] = "", "beta"
	if got := multilineEchoIndex(lines, "alpha\nbeta"); got != -1 {
		t.Fatalf("row outside prompt: echo index %d, want -1", got)
	}
	if got := multilineEchoIndex(lines, "\nalpha\n"); got != -1 {
		t.Fatalf("one nonempty command line: echo index %d, want -1", got)
	}
}

func TestMultilineEchoGreedyMatchesRestartSearch(t *testing.T) {
	// Exhaust all four-row blocks, including overlapping parts and blanks.
	choices := []string{"", "alpha", "beta", "alpha beta"}
	for mask := 0; mask < 256; mask++ {
		lines := make([]string, 4)
		value := mask
		for i := range lines {
			lines[i] = choices[value%4]
			value /= 4
		}
		for _, command := range []string{"alpha\nbeta", "alpha\nalpha", "beta\nalpha\nbeta", "\nalpha\n\nbeta\n"} {
			var parts []string
			for _, part := range strings.Split(command, "\n") {
				if part = strings.TrimSpace(part); part != "" {
					parts = append(parts, part)
				}
			}
			want := -1
			for start := range lines {
				index, matched, last := start, 0, -1
				for _, part := range parts {
					for index < len(lines) && !strings.Contains(lines[index], part) {
						index++
					}
					if index == len(lines) {
						break
					}
					matched++
					last = index
					index++
				}
				if matched == len(parts) {
					want = last
					break
				}
			}
			if got := multilineEchoIndex(lines, command); got != want {
				t.Fatalf("lines=%q command=%q got=%d want=%d", lines, command, got, want)
			}
		}
	}
}
