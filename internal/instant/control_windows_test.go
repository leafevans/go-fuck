//go:build windows

package instant

import "testing"

func TestControlCursorMovement(t *testing.T) {
	terminal := newTerminal(10, 5)
	terminal.cursorX, terminal.cursorY = 3, 3

	terminal.control("2", 'A')
	if terminal.cursorY != 1 {
		t.Fatalf("cursor up = %d, want 1", terminal.cursorY)
	}
	terminal.control("1", 'B')
	if terminal.cursorY != 2 {
		t.Fatalf("cursor down = %d, want 2", terminal.cursorY)
	}
	terminal.control("4", 'C')
	if terminal.cursorX != 7 {
		t.Fatalf("cursor right = %d, want 7", terminal.cursorX)
	}
	terminal.control("2", 'D')
	if terminal.cursorX != 5 {
		t.Fatalf("cursor left = %d, want 5", terminal.cursorX)
	}

	terminal.cursorX, terminal.cursorY = 5, 2
	terminal.control("1", 'E')
	if terminal.cursorX != 0 || terminal.cursorY != 3 {
		t.Fatalf("cursor next line = %d,%d, want 0,3", terminal.cursorX, terminal.cursorY)
	}
	terminal.control("2", 'F')
	if terminal.cursorX != 0 || terminal.cursorY != 1 {
		t.Fatalf("cursor previous line = %d,%d, want 0,1", terminal.cursorX, terminal.cursorY)
	}
	terminal.control("5", 'G')
	if terminal.cursorX != 4 {
		t.Fatalf("cursor column = %d, want 4", terminal.cursorX)
	}
	terminal.control("4", 'd')
	if terminal.cursorY != 3 {
		t.Fatalf("cursor row = %d, want 3", terminal.cursorY)
	}
	terminal.control("3;5", 'H')
	if terminal.cursorX != 4 || terminal.cursorY != 2 {
		t.Fatalf("cursor position = %d,%d, want 4,2", terminal.cursorX, terminal.cursorY)
	}
	terminal.control("2;4", 'f')
	if terminal.cursorX != 3 || terminal.cursorY != 1 {
		t.Fatalf("cursor position f = %d,%d, want 3,1", terminal.cursorX, terminal.cursorY)
	}
}

func TestControlEraseAndScrollBranches(t *testing.T) {
	terminal := newLineTerminal("abcdef", "ghijkl", "mnopqr")
	terminal.cursorX, terminal.cursorY = 2, 1
	terminal.control("2", 'K')
	if terminal.screen[1] != nil {
		t.Fatalf("erase line = %q", rowText(terminal, 1))
	}

	terminal.control("2", 'J')
	if terminal.screen[0] != nil || terminal.screen[2] != nil {
		t.Fatalf("erase display cleared %q and %q", rowText(terminal, 0), rowText(terminal, 2))
	}

	terminal = newLineTerminal("abcdef", "ghijkl", "mnopqr")
	terminal.cursorX, terminal.cursorY = 2, 0
	terminal.control("2", 'X')
	if got := rowText(terminal, 0); got != "ab  ef" {
		t.Fatalf("erase characters = %q", got)
	}
	terminal.control("2", 'P')
	if got := rowText(terminal, 0); got != "abef" {
		t.Fatalf("delete characters = %q", got)
	}
	terminal.cursorX = 2
	terminal.control("2", '@')
	if got := rowText(terminal, 0); got != "ab  ef" {
		t.Fatalf("insert characters = %q", got)
	}

	terminal = newLineTerminal("a", "b", "c")
	terminal.cursorY = 1
	terminal.control("1", 'L')
	if rowText(terminal, 1) != "" || rowText(terminal, 2) != "b" {
		t.Fatalf("insert lines = %q, %q", rowText(terminal, 1), rowText(terminal, 2))
	}
	terminal = newLineTerminal("a", "b", "c")
	terminal.cursorY = 1
	terminal.control("1", 'M')
	if rowText(terminal, 1) != "c" {
		t.Fatalf("delete lines = %q", rowText(terminal, 1))
	}

	terminal = newLineTerminal("a", "b", "c")
	terminal.control("1", 'S')
	if terminal.history[0] != "a" || rowText(terminal, 0) != "b" {
		t.Fatalf("scroll up history = %q screen = %q", terminal.history, rowText(terminal, 0))
	}
	terminal = newLineTerminal("a", "b", "c")
	terminal.control("1", 'T')
	if rowText(terminal, 0) != "" || rowText(terminal, 1) != "a" {
		t.Fatalf("scroll down = %q, %q", rowText(terminal, 0), rowText(terminal, 1))
	}
}

func TestControlPrivateModes(t *testing.T) {
	terminal := newTerminal(10, 3)
	terminal.control("?47", 'h')
	if terminal.alternate == nil {
		t.Fatal("private mode 47h did not enter the alternate screen")
	}
	terminal.control("?47", 'l')
	if terminal.alternate != nil {
		t.Fatal("private mode 47l did not leave the alternate screen")
	}
	terminal.control("?25", 'h')
	if terminal.alternate != nil {
		t.Fatal("an unsupported private mode changed the screen")
	}
	terminal.control("2", 'h')
	if terminal.alternate != nil {
		t.Fatal("a non-private mode changed the screen")
	}
}

func TestEscapeSequences(t *testing.T) {
	terminal := newTerminal(10, 3)
	if got := terminal.escape("\x1b[2K", 0); got != 4 {
		t.Fatalf("escape CSI = %d, want 4", got)
	}
	if got := terminal.escape("\x1b]0;title\x07", 0); got != len("\x1b]0;title\x07") {
		t.Fatalf("escape OSC BEL = %d", got)
	}
	if got := terminal.escape("\x1b]0;title\x1b\\", 0); got != len("\x1b]0;title\x1b\\") {
		t.Fatalf("escape OSC ST = %d", got)
	}
	if got := terminal.escape("\x1b]unterminated", 0); got != len("\x1b]unterminated") {
		t.Fatalf("escape unterminated OSC = %d", got)
	}

	terminal.cursorX, terminal.cursorY = 2, 1
	if got := terminal.escape("\x1b7", 0); got != 2 {
		t.Fatalf("escape save cursor = %d", got)
	}
	terminal.cursorX, terminal.cursorY = 0, 0
	if got := terminal.escape("\x1b8", 0); got != 2 {
		t.Fatalf("escape restore cursor = %d", got)
	}
	if terminal.cursorX != 2 || terminal.cursorY != 1 {
		t.Fatalf("restored cursor = %d,%d, want 2,1", terminal.cursorX, terminal.cursorY)
	}

	terminal.cursorY = 1
	terminal.escape("\x1bM", 0)
	if terminal.cursorY != 0 {
		t.Fatalf("escape reverse index = %d, want 0", terminal.cursorY)
	}
	terminal.escape("\x1bM", 0)
	if rowText(terminal, 0) != "" {
		t.Fatalf("escape reverse index at the top did not scroll: %q", rowText(terminal, 0))
	}

	terminal = newLineTerminal("a", "b")
	terminal.escape("\x1bc", 0)
	if terminal.screen[0] != nil || terminal.cursorX != 0 || terminal.cursorY != 0 {
		t.Fatalf("escape reset = %q, %d,%d", rowText(terminal, 0), terminal.cursorX, terminal.cursorY)
	}
	if got := terminal.escape("\x1bZ", 0); got != 2 {
		t.Fatalf("escape unknown = %d, want 2", got)
	}
	if got := terminal.escape("\x1b", 0); got != 1 {
		t.Fatalf("escape truncated = %d, want 1", got)
	}
	if got := terminal.escape("\x1b[", 0); got != 2 {
		t.Fatalf("escape truncated CSI = %d, want 2", got)
	}
}

func TestClampAndDefaults(t *testing.T) {
	if got := clamp(-1, 0, 10); got != 0 {
		t.Fatalf("clamp below = %d", got)
	}
	if got := clamp(11, 0, 10); got != 10 {
		t.Fatalf("clamp above = %d", got)
	}
	if got := clamp(5, 0, 10); got != 5 {
		t.Fatalf("clamp inside = %d", got)
	}

	terminal := newTerminal(0, 0)
	if terminal.columns != defaultRenderColumns || len(terminal.screen) != defaultRenderRows {
		t.Fatalf("newTerminal defaults = %d columns, %d rows", terminal.columns, len(terminal.screen))
	}
}

func TestRuneWidth(t *testing.T) {
	cases := map[rune]int{
		'a':      1,
		'中':      2,
		'😀':      2,
		'\u0301': 0,
		'\u200b': 1,
		'\x01':   0,
	}
	for character, want := range cases {
		if got := runeWidth(character); got != want {
			t.Fatalf("runeWidth(%q) = %d, want %d", character, got, want)
		}
	}
}
