//go:build windows

package instant

import "testing"

// newLineTerminal returns a terminal with the given screen lines.
func newLineTerminal(lines ...string) *terminal {
	terminal := newTerminal(20, len(lines))
	for index, line := range lines {
		terminal.screen[index] = screenLine([]rune(line))
	}
	return terminal
}

// rowText returns the text of one screen row.
func rowText(terminal *terminal, row int) string {
	return string(terminal.screen[row])
}

func TestEraseCharacters(t *testing.T) {
	terminal := newLineTerminal("abcdef")
	terminal.cursorX = 2
	terminal.eraseCharacters(2)
	if got := rowText(terminal, 0); got != "ab  ef" {
		t.Fatalf("eraseCharacters() = %q, want %q", got, "ab  ef")
	}
	terminal.eraseCharacters(0)
	if got := rowText(terminal, 0); got != "ab  ef" {
		t.Fatalf("eraseCharacters(0) = %q, want no change", got)
	}
}

func TestDeleteAndInsertCharacters(t *testing.T) {
	terminal := newLineTerminal("abcdef")
	terminal.cursorX = 2
	terminal.deleteCharacters(2)
	if got := rowText(terminal, 0); got != "abef" {
		t.Fatalf("deleteCharacters() = %q, want %q", got, "abef")
	}
	terminal.cursorX = 2
	terminal.insertCharacters(2)
	if got := rowText(terminal, 0); got != "ab  ef" {
		t.Fatalf("insertCharacters() = %q, want %q", got, "ab  ef")
	}
}

func TestBlankAndTruncateLine(t *testing.T) {
	terminal := newLineTerminal("abcdef")
	terminal.blankLine(0, 2)
	if got := rowText(terminal, 0); got != "   def" {
		t.Fatalf("blankLine() = %q, want %q", got, "   def")
	}
	terminal = newLineTerminal("abcdef")
	terminal.truncateLine(0, 2)
	if got := rowText(terminal, 0); got != "ab" {
		t.Fatalf("truncateLine() = %q, want %q", got, "ab")
	}
	// Out of range rows are ignored.
	terminal.blankLine(9, 2)
	terminal.truncateLine(9, 2)
}

func TestEraseLine(t *testing.T) {
	terminal := newLineTerminal("abcdef", "ghijkl")
	terminal.cursorX = 3
	terminal.cursorY = 0
	terminal.eraseLine(0)
	if got := rowText(terminal, 0); got != "abc" {
		t.Fatalf("eraseLine(0) = %q, want %q", got, "abc")
	}
	terminal.cursorY = 1
	terminal.eraseLine(1)
	if got := rowText(terminal, 1); got != "    kl" {
		t.Fatalf("eraseLine(1) = %q, want %q", got, "    kl")
	}
	terminal.eraseLine(2)
	if terminal.screen[1] != nil {
		t.Fatalf("eraseLine(2) left %q, want nil", rowText(terminal, 1))
	}
}

func TestEraseDisplay(t *testing.T) {
	terminal := newLineTerminal("abcdef", "ghijkl", "mnopqr")
	terminal.cursorX = 3
	terminal.cursorY = 1
	terminal.eraseDisplay(0)
	if got := rowText(terminal, 0); got != "abcdef" {
		t.Fatalf("eraseDisplay(0) row 0 = %q, want unchanged", got)
	}
	if got := rowText(terminal, 1); got != "ghi" {
		t.Fatalf("eraseDisplay(0) row 1 = %q, want %q", got, "ghi")
	}
	if terminal.screen[2] != nil {
		t.Fatalf("eraseDisplay(0) row 2 = %q, want nil", rowText(terminal, 2))
	}

	terminal = newLineTerminal("abcdef", "ghijkl")
	terminal.cursorX = 2
	terminal.cursorY = 1
	terminal.eraseDisplay(1)
	if terminal.screen[0] != nil {
		t.Fatalf("eraseDisplay(1) row 0 = %q, want nil", rowText(terminal, 0))
	}
	if got := rowText(terminal, 1); got != "   jkl" {
		t.Fatalf("eraseDisplay(1) row 1 = %q, want %q", got, "   jkl")
	}

	terminal = newLineTerminal("abc", "def")
	terminal.eraseDisplay(2)
	for row, line := range terminal.screen {
		if line != nil {
			t.Fatalf("eraseDisplay(2) row %d = %q, want nil", row, rowText(terminal, row))
		}
	}
}

func TestInsertAndDeleteLines(t *testing.T) {
	terminal := newLineTerminal("a", "b", "c")
	terminal.cursorY = 1
	terminal.insertLines(1)
	if rowText(terminal, 0) != "a" || rowText(terminal, 1) != "" || rowText(terminal, 2) != "b" {
		t.Fatalf("insertLines() = %q, %q, %q", rowText(terminal, 0), rowText(terminal, 1), rowText(terminal, 2))
	}

	terminal = newLineTerminal("a", "b", "c")
	terminal.cursorY = 1
	terminal.deleteLines(1)
	if rowText(terminal, 0) != "a" || rowText(terminal, 1) != "c" || rowText(terminal, 2) != "" {
		t.Fatalf("deleteLines() = %q, %q, %q", rowText(terminal, 0), rowText(terminal, 1), rowText(terminal, 2))
	}
}

func TestScrollUpAndDown(t *testing.T) {
	terminal := newLineTerminal("a", "b", "c")
	terminal.scrollUp(1)
	if len(terminal.history) != 1 || terminal.history[0] != "a" {
		t.Fatalf("scrollUp() history = %q, want [a]", terminal.history)
	}
	if rowText(terminal, 0) != "b" || rowText(terminal, 1) != "c" || rowText(terminal, 2) != "" {
		t.Fatalf("scrollUp() screen = %q, %q, %q", rowText(terminal, 0), rowText(terminal, 1), rowText(terminal, 2))
	}

	terminal = newLineTerminal("a", "b", "c")
	terminal.scrollDown(1)
	if rowText(terminal, 0) != "" || rowText(terminal, 1) != "a" || rowText(terminal, 2) != "b" {
		t.Fatalf("scrollDown() screen = %q, %q, %q", rowText(terminal, 0), rowText(terminal, 1), rowText(terminal, 2))
	}
}

func TestTab(t *testing.T) {
	terminal := newTerminal(20, 2)
	terminal.cursorX = 2
	terminal.tab()
	if terminal.cursorX != 8 {
		t.Fatalf("tab() cursorX = %d, want 8", terminal.cursorX)
	}
}
