package instant

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/leafevans/go-fuck/internal/consts"
)

// prompt returns a log line with the instant mode prompt marker.
func prompt(text string) string {
	return consts.UserCommandMark + text
}

// writeLog writes a terminal log fixture to a temporary file.
func writeLog(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "log")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write log: %v", err)
	}
	return path
}

func TestLastOutput(t *testing.T) {
	log := prompt("$ git brnch") + "\r\n" +
		"git: 'brnch' is not a git command.\r\n" +
		prompt("$ ") + "\r\n"

	got, ok := LastOutput(writeLog(t, log), "git brnch")
	if !ok {
		t.Fatal("LastOutput() did not find the command")
	}
	if want := "git: 'brnch' is not a git command."; got != want {
		t.Fatalf("LastOutput() = %q, want %q", got, want)
	}
}

func TestLastOutputStripsEscapeSequences(t *testing.T) {
	log := prompt("$ ls") + "\r\n" +
		"\x1b[31mfile one\x1b[0m\r\n" +
		"file two\x07\r\n"
	got, ok := LastOutput(writeLog(t, log), "ls")
	if !ok {
		t.Fatal("LastOutput() did not find the command")
	}
	if want := "file one\nfile two"; got != want {
		t.Fatalf("LastOutput() = %q, want %q", got, want)
	}
}

func TestLastOutputUsesLastCall(t *testing.T) {
	log := prompt("$ echo old") + "\r\nold\r\n" +
		prompt("$ echo new") + "\r\nnew\r\n" +
		prompt("$ ")
	got, ok := LastOutput(writeLog(t, log), "echo new")
	if !ok || got != "new" {
		t.Fatalf("LastOutput() = %q, %t, want new, true", got, ok)
	}

	if _, ok := LastOutput(writeLog(t, log), "git brnch"); ok {
		t.Fatal("LastOutput() should not find a command that is not in the log")
	}
}

func TestLastOutputHandlesCarriageReturn(t *testing.T) {
	log := prompt("$ make") + "\r\n" +
		"progress 1\rprogress 100\r\n" +
		prompt("$ ")
	got, ok := LastOutput(writeLog(t, log), "make")
	if !ok || got != "progress 100" {
		t.Fatalf("LastOutput() = %q, %t, want progress 100, true", got, ok)
	}
}

func TestLastOutputEmptyOutput(t *testing.T) {
	log := prompt("$ true") + "\r\n" + prompt("$ ")
	got, ok := LastOutput(writeLog(t, log), "true")
	if !ok {
		t.Fatal("LastOutput() did not find the command")
	}
	if got != "" {
		t.Fatalf("LastOutput() = %q, want empty output", got)
	}
}

func TestLastOutputMissingLog(t *testing.T) {
	if _, ok := LastOutput(filepath.Join(t.TempDir(), "missing"), "ls"); ok {
		t.Fatal("LastOutput() should fail for a missing log file")
	}
}

func TestLogPath(t *testing.T) {
	t.Setenv(consts.OutputLogEnv, "")
	t.Setenv(consts.LegacyOutputLogEnv, "")
	if path := LogPath(); path != "" {
		t.Fatalf("LogPath() = %q, want empty", path)
	}

	t.Setenv(consts.LegacyOutputLogEnv, "/tmp/legacy.log")
	if path := LogPath(); path != "/tmp/legacy.log" {
		t.Fatalf("LogPath() = %q, want the legacy path", path)
	}

	t.Setenv(consts.OutputLogEnv, "/tmp/go-fuck.log")
	if path := LogPath(); path != "/tmp/go-fuck.log" {
		t.Fatalf("LogPath() = %q, want the modern path", path)
	}
}

func TestRingWriterTrimsOldBytes(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ring.log")
	file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0o600)
	if err != nil {
		t.Fatalf("open log: %v", err)
	}
	defer file.Close()

	writer := NewRingWriter(file)
	old := bytes.Repeat([]byte("a"), consts.LogSizeInBytes)
	if _, err := writer.Write(old); err != nil {
		t.Fatalf("write old data: %v", err)
	}
	if _, err := writer.Write([]byte("TAIL")); err != nil {
		t.Fatalf("write tail data: %v", err)
	}

	info, err := file.Stat()
	if err != nil {
		t.Fatalf("stat log: %v", err)
	}
	if info.Size() > consts.LogSizeInBytes {
		t.Fatalf("log size = %d, want at most %d", info.Size(), consts.LogSizeInBytes)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read log: %v", err)
	}
	if !bytes.HasSuffix(data, []byte("TAIL")) {
		t.Fatalf("log does not end with the newest bytes: %q", string(data[len(data)-10:]))
	}
	if len(data) >= consts.LogSizeInBytes {
		t.Fatalf("the log was not trimmed: %d bytes", len(data))
	}
	if bytes.Count(data, []byte("a")) >= consts.LogSizeInBytes {
		t.Fatal("the old bytes were not dropped from the beginning")
	}
}

func TestRingWriterNilFile(t *testing.T) {
	writer := NewRingWriter(nil)
	if n, err := writer.Write([]byte("data")); err != nil || n != 4 {
		t.Fatalf("Write() = %d, %v, want 4, nil", n, err)
	}
}

func TestSanitizeRemovesMarkerAndControls(t *testing.T) {
	input := prompt("\x1b[32mok\x1b[0m\b!")
	got := sanitize(input)
	if strings.Contains(got, consts.UserCommandMark) || strings.Contains(got, "\x1b") {
		t.Fatalf("sanitize() left control data: %q", got)
	}
	if want := "ok!"; got != want {
		t.Fatalf("sanitize() = %q, want %q", got, want)
	}
}
