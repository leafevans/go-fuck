//go:build windows

package instant

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/leafevans/go-fuck/internal/consts"
)

// writeRenderedLog writes a ConPTY terminal log fixture to a temporary file.
func writeRenderedLog(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "rendered.log")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write log: %v", err)
	}
	return path
}

// sizeRecord returns the size record the Windows shell logger writes.
func sizeRecord(columns, rows int) string {
	return conPTYSizeRecord + " " + itoa(columns) + " " + itoa(rows) + "\n"
}

func itoa(value int) string {
	if value <= 0 {
		return "0"
	}
	var digits []byte
	for value > 0 {
		digits = append([]byte{byte('0' + value%10)}, digits...)
		value /= 10
	}
	return string(digits)
}

// TestLogLinesRendersCursorPositioning checks that ConPTY output which encodes
// line breaks as absolute cursor positions is rendered into separate lines.
func TestLogLinesRendersCursorPositioning(t *testing.T) {
	data := sizeRecord(120, 30) +
		consts.UserCommandMark + "PS> git brnch\r\n" +
		"git: 'brnch' is not a git command. See 'git --help'." +
		"\x1b[?25l\x1b[4;1HThe most similar command is\x1b[5;9Hbranch\r\n" +
		consts.UserCommandMark + "PS> "

	path := writeRenderedLog(t, data)
	output, ok := LastOutput(path, "git brnch")
	if !ok {
		t.Fatal("LastOutput() did not find the command")
	}
	for _, want := range []string{"git: 'brnch' is not a git command.", "The most similar command is", "branch"} {
		if !strings.Contains(output, want) {
			t.Fatalf("LastOutput() = %q, want it to contain %q", output, want)
		}
	}
}

// TestLogLinesScrollsIntoTheScrollback checks that content which leaves the
// viewport is kept in the rendered output.
func TestLogLinesScrollsIntoTheScrollback(t *testing.T) {
	data := sizeRecord(40, 4) + "one\r\ntwo\r\nthree\r\nfour\r\nfive\r\n"
	lines := logLines(data)
	// The initial viewport is taller than the size record, so the empty rows
	// end up at the beginning of the scrollback.
	start := 0
	for start < len(lines) && lines[start] == "" {
		start++
	}
	want := []string{"one", "two", "three", "four", "five"}
	if len(lines)-start < len(want) {
		t.Fatalf("logLines() = %q, want at least %d non empty lines after %d empty lines", lines, len(want), start)
	}
	for index, expected := range want {
		if lines[start+index] != expected {
			t.Fatalf("logLines()[%d] = %q, want %q (all: %q)", start+index, lines[start+index], expected, lines)
		}
	}
}

// TestLogLinesCarriageReturnOverwrite checks the classic progress bar pattern.
func TestLogLinesCarriageReturnOverwrite(t *testing.T) {
	data := sizeRecord(80, 30) + "progress 1\rprogress 100\r\n"
	lines := logLines(data)
	if len(lines) == 0 || lines[0] != "progress 100" {
		t.Fatalf("logLines() = %q, want the overwritten line", lines)
	}
}

// TestLogLinesErasesTheLine checks EL sequences, which PSReadLine uses while
// it redraws the command line.
func TestLogLinesErasesTheLine(t *testing.T) {
	data := sizeRecord(80, 30) + "old text\rnew\x1b[K\r\n"
	lines := logLines(data)
	if len(lines) == 0 || lines[0] != "new" {
		t.Fatalf("logLines() = %q, want the erased line", lines)
	}
}

// TestLogLinesHandlesResize checks that a resize record updates the viewport.
func TestLogLinesHandlesResize(t *testing.T) {
	data := sizeRecord(40, 4) + "one\r\ntwo\r\nthree\r\nfour\r\n" +
		sizeRecord(40, 2) + "five\r\nsix\r\n"
	lines := logLines(data)
	joined := strings.Join(lines, "\n")
	for _, want := range []string{"one", "two", "three", "four", "five", "six"} {
		if !strings.Contains(joined, want) {
			t.Fatalf("logLines() = %q, want it to contain %q", lines, want)
		}
	}
}

// TestLogLinesSkipsOSC checks that window title sequences are not rendered.
func TestLogLinesSkipsOSC(t *testing.T) {
	title := string([]byte{27}) + "]0;window title" + string([]byte{7})
	data := sizeRecord(80, 30) + title + "file one\r\n"
	lines := logLines(data)
	if len(lines) == 0 || lines[0] != "file one" {
		t.Fatalf("logLines() = %q, want file one", lines)
	}
}
