//go:build linux

package instant

import (
	"io"
	"os"
	"testing"
)

// openTestTerminal opens a pseudo terminal pair for the terminal tests.
func openTestTerminal(t *testing.T) (*os.File, *os.File) {
	t.Helper()
	master, slave, err := openPTY()
	if err != nil {
		t.Fatalf("openPTY() error = %v", err)
	}
	t.Cleanup(func() {
		_ = master.Close()
		_ = slave.Close()
	})
	return master, slave
}

// TestTermiosRoundTrip covers getTermios and setTermios.
func TestTermiosRoundTrip(t *testing.T) {
	_, slave := openTestTerminal(t)
	state, ok := getTermios(slave.Fd())
	if !ok {
		t.Fatal("getTermios() = false for a pseudo terminal")
	}
	if err := setTermios(slave.Fd(), state); err != nil {
		t.Fatalf("setTermios() error = %v", err)
	}
}

// TestWindowSizeOnTerminal covers the success path of getWindowSize and
// setWindowSize.
func TestWindowSizeOnTerminal(t *testing.T) {
	master, _ := openTestTerminal(t)
	if _, ok := getWindowSize(master.Fd()); !ok {
		t.Fatal("getWindowSize() = false for a pseudo terminal")
	}
	want := windowSize{Rows: 32, Cols: 100}
	if err := setWindowSize(master.Fd(), want); err != nil {
		t.Fatalf("setWindowSize() error = %v", err)
	}
	got, ok := getWindowSize(master.Fd())
	if !ok {
		t.Fatal("getWindowSize() = false after setWindowSize()")
	}
	if got.Rows != want.Rows || got.Cols != want.Cols {
		t.Fatalf("getWindowSize() = %+v, want %+v", got, want)
	}
}

// TestMakeRawTerminalOnTerminal covers the raw mode and restore path of
// makeRawTerminal.
func TestMakeRawTerminalOnTerminal(t *testing.T) {
	_, slave := openTestTerminal(t)
	restore := makeRawTerminal(slave)
	if restore == nil {
		t.Fatal("makeRawTerminal() = nil for a pseudo terminal")
	}
	restore()
}

// TestRunLoggedShellWithTerminalStdin covers the terminal restore and window
// size branches of runLoggedShell.
func TestRunLoggedShellWithTerminalStdin(t *testing.T) {
	_, slave := openTestTerminal(t)
	previous := os.Stdin
	os.Stdin = slave
	t.Cleanup(func() { os.Stdin = previous })

	code, err := runLoggedShell(lookupTool(t, "true"), slave, io.Discard, io.Discard)
	if err != nil {
		t.Fatalf("runLoggedShell() error = %v", err)
	}
	if code != 0 {
		t.Fatalf("runLoggedShell() = %d, want 0", code)
	}
}
