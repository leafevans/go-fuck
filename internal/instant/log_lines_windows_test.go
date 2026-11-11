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

// TestLogLinesKeepsTheMarkerAcrossPromptRedraws checks that the instant mode
// prompt marker survives the erase and redraw sequences that PSReadLine writes
// while it edits the command line.
func TestLogLinesKeepsTheMarkerAcrossPromptRedraws(t *testing.T) {
	data := sizeRecord(80, 30) +
		consts.UserCommandMark + "PS> git brnch" + "\x1b[K" + "\r\x1b[K" +
		consts.UserCommandMark + "PS> git brnch\r\n" +
		"git: 'brnch' is not a git command. See 'git --help'.\r\n" +
		consts.UserCommandMark + "PS> "

	output, ok := LastOutput(writeRenderedLog(t, data), "git brnch")
	if !ok {
		t.Fatal("LastOutput() did not find the command")
	}
	if want := "git: 'brnch' is not a git command. See 'git --help'."; output != want {
		t.Fatalf("LastOutput() = %q, want %q", output, want)
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

// renderedNonEmptyLines returns the non empty lines of a terminal log.
func renderedNonEmptyLines(data string) []string {
	nonEmpty := []string{}
	for _, line := range logLines(data) {
		if line != "" {
			nonEmpty = append(nonEmpty, line)
		}
	}
	return nonEmpty
}

// TestLogLinesRendersWideCharacters checks that CJK text is not mangled.
func TestLogLinesRendersWideCharacters(t *testing.T) {
	data := sizeRecord(80, 30) + "错误: 未找到命令\r\n"
	lines := renderedNonEmptyLines(data)
	if len(lines) == 0 || lines[0] != "错误: 未找到命令" {
		t.Fatalf("logLines() = %q, want the CJK line", lines)
	}
}

// TestLogLinesWideCharacterCursorPositioning checks that cursor positions after
// wide characters match the columns the terminal actually used.
func TestLogLinesWideCharacterCursorPositioning(t *testing.T) {
	data := sizeRecord(80, 30) + "中文\x1b[5GOK\r\n"
	lines := renderedNonEmptyLines(data)
	if len(lines) == 0 || lines[0] != "中文OK" {
		t.Fatalf("logLines() = %q, want the wide character line with the positioned text", lines)
	}
}

// TestLogLinesJoinsWrappedWideCharacters checks that a wide character which
// does not fit into the remaining columns starts the next row and that the
// wrapped rows are joined again into one logical line.
func TestLogLinesJoinsWrappedWideCharacters(t *testing.T) {
	data := sizeRecord(5, 3) + "中文中文"
	lines := renderedNonEmptyLines(data)
	want := []string{"中文中文"}
	if len(lines) != len(want) {
		t.Fatalf("logLines() = %q, want %q", lines, want)
	}
	for index, expected := range want {
		if lines[index] != expected {
			t.Fatalf("logLines()[%d] = %q, want %q (all: %q)", index, lines[index], expected, lines)
		}
	}
}

// TestLogLinesJoinsWrappedOutputRows checks that ConPTY rows which continue on
// the next row are joined again. ConPTY emits the wrapped text of the child
// console with an explicit line break and leaves the last column empty, so the
// rules see the logical line of an error message instead of terminal rows.
func TestLogLinesJoinsWrappedOutputRows(t *testing.T) {
	data := sizeRecord(20, 30) + "0123456789012345678\r\ncontinues\r\n"
	lines := renderedNonEmptyLines(data)
	want := "0123456789012345678continues"
	if len(lines) != 1 || lines[0] != want {
		t.Fatalf("logLines() = %q, want %q", lines, want)
	}
}

// TestLogLinesOverwritesHalfOfAWideCharacter checks that a cursor inside a
// wide character clears the whole character before writing.
func TestLogLinesOverwritesHalfOfAWideCharacter(t *testing.T) {
	data := sizeRecord(80, 30) + "中文\x1b[2GX\r\n"
	lines := renderedNonEmptyLines(data)
	if len(lines) == 0 || lines[0] != " X文" {
		t.Fatalf("logLines() = %q, want the repaired wide character line", lines)
	}
}

// TestLogLinesRendersEmoji checks that emoji occupy two columns as well.
func TestLogLinesRendersEmoji(t *testing.T) {
	data := sizeRecord(80, 30) + "🎉ok\r\n"
	lines := renderedNonEmptyLines(data)
	if len(lines) == 0 || lines[0] != "🎉ok" {
		t.Fatalf("logLines() = %q, want the emoji line", lines)
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

// TestLogLinesWrappedCommand checks that a command that is wider than the
// terminal and therefore wrapped by the screen model is still found.
func TestLogLinesWrappedCommand(t *testing.T) {
	command := "Write-Output averylongargumentname"
	data := sizeRecord(20, 30) +
		consts.UserCommandMark + "PS> " + command + "\r\n" +
		"done\r\n" +
		consts.UserCommandMark + "PS> "

	output, ok := LastOutput(writeRenderedLog(t, data), command)
	if !ok {
		t.Fatal("LastOutput() did not find the wrapped command")
	}
	if want := "done"; output != want {
		t.Fatalf("LastOutput() = %q, want %q", output, want)
	}
}

// TestLogLinesIgnoresAlternateScreenContent checks that a full screen program
// (vim, tmux, ...) does not leak its screen into the log and that the previous
// screen is restored afterwards.
func TestLogLinesIgnoresAlternateScreenContent(t *testing.T) {
	data := sizeRecord(80, 30) +
		consts.UserCommandMark + "PS> vim file.txt\r\n" +
		"before\r\n" +
		"\x1b[?1049h" +
		"VIM: full screen content\r\n" +
		"\x1b[?1049l" +
		"after\r\n" +
		consts.UserCommandMark + "PS> "

	lines := logLines(data)
	joined := strings.Join(lines, "\n")
	if strings.Contains(joined, "VIM: full screen content") {
		t.Fatalf("logLines() kept the alternate screen content: %q", lines)
	}
	for _, want := range []string{"before", "after"} {
		if !strings.Contains(joined, want) {
			t.Fatalf("logLines() = %q, want it to contain %q", lines, want)
		}
	}
}
