//go:build windows

package system

import "testing"

func TestConsoleModeHelpers(t *testing.T) {
	if _, ok := getConsoleMode(0); ok {
		t.Fatal("getConsoleMode(0) succeeded")
	}
	if setConsoleMode(0, 0) {
		t.Fatal("setConsoleMode(0) succeeded")
	}
	_ = stdHandle(stdinHandle)
	InitOutput()
	RestoreInput()
	initInput()
	RestoreInput()
}

func TestRestoreInputClearsSavedState(t *testing.T) {
	previousSaved := inputModeSaved
	previousMode := savedInputMode
	t.Cleanup(func() {
		inputModeSaved = previousSaved
		savedInputMode = previousMode
	})

	inputModeSaved = true
	savedInputMode = enableProcessedInput | enableLineInput | enableEchoInput
	RestoreInput()
	if inputModeSaved {
		t.Fatal("RestoreInput() left inputModeSaved set")
	}
}
