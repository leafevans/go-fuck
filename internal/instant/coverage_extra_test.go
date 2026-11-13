package instant

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/leafevans/go-fuck/v2/internal/consts"
)

// TestLastOutputWhitespaceScript covers the empty parts branch of echoIndex.
func TestLastOutputWhitespaceScript(t *testing.T) {
	log := prompt("$ git brnch") + "\r\n" +
		"git: 'brnch' is not a git command.\r\n" +
		prompt("$ ")
	if _, ok := LastOutput(writeLog(t, log), "   "); ok {
		t.Fatal("LastOutput() found a whitespace only command")
	}
}

// TestEchoIndexMatchesSeparateParts covers the branch that matches a command
// whose arguments appear in a line separated by extra whitespace.
func TestEchoIndexMatchesSeparateParts(t *testing.T) {
	call := logCall{prompt: prompt("$ git  brnch")}
	if got := call.echoIndex("git brnch", []string{"git", "brnch"}); got != 0 {
		t.Fatalf("echoIndex() = %d, want 0", got)
	}
}

// TestWrappedEchoIndexLineLimit covers the prompt block line limit of
// wrappedEchoIndex.
func TestWrappedEchoIndexLineLimit(t *testing.T) {
	log := prompt("$ ") + "\r\n"
	for index := 0; index < 10; index++ {
		log += "output line" + "\r\n"
	}
	if _, ok := LastOutput(writeLog(t, log), "totally-missing-command"); ok {
		t.Fatal("LastOutput() found a command that is not in the log")
	}
}

// TestWrappedEchoIndexCommandNotAtLineEnd covers the branch that rejects a
// wrapped command when the line continues after it.
func TestWrappedEchoIndexCommandNotAtLineEnd(t *testing.T) {
	call := logCall{
		prompt: prompt("$ "),
		lines:  []string{"git chec", "kout extra"},
	}
	if got := call.echoIndex("git checkout", []string{"git", "checkout"}); got != -1 {
		t.Fatalf("echoIndex() = %d, want -1", got)
	}
}

// TestLogCallOutputRendersCarriageReturnAndTrimsBlankLines covers the carriage
// return overwrite and the trailing blank line trimming of logCall.output.
func TestLogCallOutputRendersCarriageReturnAndTrimsBlankLines(t *testing.T) {
	call := logCall{
		prompt: prompt("$ make"),
		lines:  []string{"progress 1\rprogress 100", "done", "", ""},
	}
	if got := call.output(0); got != "progress 100\ndone" {
		t.Fatalf("output() = %q, want %q", got, "progress 100\ndone")
	}
}

// TestRingWriterStatFailure covers the stat error branch of RingWriter.Write.
func TestRingWriterStatFailure(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ring.log")
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		t.Fatalf("open log: %v", err)
	}
	writer := NewRingWriter(file)
	if err := file.Close(); err != nil {
		t.Fatalf("close log: %v", err)
	}
	if _, err := writer.Write([]byte("data")); err == nil {
		t.Fatal("Write() of a closed file returned no error")
	}
}

// TestRingWriterTrimFailure covers the branches of Write that clamp the kept
// bytes and report a failed trimming pass.
func TestRingWriterTrimFailure(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ring.log")
	if err := os.WriteFile(path, []byte("old"), 0o600); err != nil {
		t.Fatal(err)
	}
	file, err := os.Open(path)
	if err != nil {
		t.Fatalf("open log: %v", err)
	}
	defer file.Close()

	writer := NewRingWriter(file)
	data := make([]byte, consts.LogSizeInBytes+1)
	if _, err := writer.Write(data); err == nil {
		if runtime.GOOS != "windows" {
			t.Skip("the read only log could be truncated (running as root?)")
		}
		t.Fatal("Write() of a read only log returned no error")
	}
}
