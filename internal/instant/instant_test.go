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

func TestLastOutputStripsOSC(t *testing.T) {
	osc := string([]byte{27}) + "]0;window title" + string([]byte{7})
	log := prompt("$ ls") + "\r\n" + osc + "file one" + "\r\n"
	got, ok := LastOutput(writeLog(t, log), "ls")
	if !ok {
		t.Fatal("LastOutput() did not find the command")
	}
	if want := "file one"; got != want {
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

func TestLastOutputMultilinePrompt(t *testing.T) {
	log := prompt("╭─ user@host ~") + "\r\n" +
		"╰─$ g\bgit brnch" + "\r\n" +
		"\r\n" +
		"git: 'brnch' is not a git command." + "\r\n" +
		prompt("╰─$ ")
	got, ok := LastOutput(writeLog(t, log), "git brnch")
	if !ok {
		t.Fatal("LastOutput() did not find the command in a multi line prompt")
	}
	if want := "git: 'brnch' is not a git command."; got != want {
		t.Fatalf("LastOutput() = %q, want %q", got, want)
	}
}

func TestLastOutputPowerShellContinuation(t *testing.T) {
	// A command continued with a trailing backtick is echoed over the main and
	// the continuation prompt of PowerShell.
	log := prompt("PS> Get-Content `") + "\r\n" +
		">> missing.tx\r\n" +
		">> \r\n" +
		"Get-Content : Cannot find path 'C:\\tmp\\missing.tx' because it does not exist.\r\n" +
		prompt("PS> ")

	got, ok := LastOutput(writeLog(t, log), "Get-Content `\nmissing.tx")
	if !ok {
		t.Fatal("LastOutput() did not find a backtick continued command")
	}
	if want := "Cannot find path"; !strings.Contains(got, want) {
		t.Fatalf("LastOutput() = %q, want it to contain %q", got, want)
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

func TestWrappedEchoIndex(t *testing.T) {
	call := logCall{prompt: prompt("$ git"), lines: []string{" push"}}
	if got := call.echoIndex("git push", []string{"git", "push"}); got != 1 {
		t.Fatalf("echoIndex() = %d, want 1", got)
	}
}

func TestGroupCallsTrimsCarriageReturn(t *testing.T) {
	calls := groupCalls([]string{prompt("$ ") + "\r", "output\r"})
	if len(calls) != 1 {
		t.Fatalf("groupCalls() returned %d calls, want 1", len(calls))
	}
	if strings.Contains(calls[0].prompt, "\r") {
		t.Fatalf("prompt still contains a carriage return: %q", calls[0].prompt)
	}
}

func TestStripAlternateScreen(t *testing.T) {
	input := "before\n" + "\x1b[?1049h" + "full screen\n" + "\x1b[?1049l" + "after\n"
	got := stripAlternateScreen(input)
	if strings.Contains(got, "full screen") {
		t.Fatalf("stripAlternateScreen() kept the alternate screen content: %q", got)
	}
	if !strings.Contains(got, "before") || !strings.Contains(got, "after") {
		t.Fatalf("stripAlternateScreen() = %q, want before and after", got)
	}
	if got := stripAlternateScreen("plain text\n"); got != "plain text\n" {
		t.Fatalf("stripAlternateScreen() = %q, want the input unchanged", got)
	}
}

func TestLogLinesDropAlternateScreenContent(t *testing.T) {
	data := prompt("$ vim file.txt") + "\n" +
		"before\n" +
		"\x1b[?1049h" + "VIM: full screen content\n" + "\x1b[?1049l" +
		"after\n" + prompt("$ ")
	joined := strings.Join(logLines(data), "\n")
	if strings.Contains(joined, "VIM: full screen content") {
		t.Fatalf("logLines() kept the alternate screen content: %q", joined)
	}
	for _, want := range []string{"before", "after"} {
		if !strings.Contains(joined, want) {
			t.Fatalf("logLines() = %q, want it to contain %q", joined, want)
		}
	}
}

func TestLastOutputMultilineCommand(t *testing.T) {
	log := prompt("$ Get-ChildItem |") + "\r\n" +
		"    Select-Object Name\r\n" +
		"output line\r\n" +
		prompt("$ ")

	got, ok := LastOutput(writeLog(t, log), "Get-ChildItem |\n    Select-Object Name")
	if !ok {
		t.Fatal("LastOutput() did not find the multi line command")
	}
	if want := "output line"; got != want {
		t.Fatalf("LastOutput() = %q, want %q", got, want)
	}
}

func TestLastOutputMultilineCommandFallsBackToNextCommand(t *testing.T) {
	log := prompt("$ Get-ChildItem |") + "\r\n" +
		"    Select-Object Name\r\n" +
		"output line\r\n" +
		prompt("$ ")

	if _, ok := LastOutput(writeLog(t, log), "Get-ChildItem -Path ."); ok {
		t.Fatal("LastOutput() found a different multi line command")
	}
}
