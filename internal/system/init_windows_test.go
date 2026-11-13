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
