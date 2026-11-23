package instant

import (
	"bytes"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/leafevans/go-fuck/v3/internal/consts"
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
