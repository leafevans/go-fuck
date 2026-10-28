//go:build windows

package system

import (
	"sync"
	"syscall"
	"unsafe"
)

// Windows console mode flags.
const (
	stdinHandle                     = ^uintptr(9)  // -10
	stdoutHandle                    = ^uintptr(11) // -11
	enableProcessedInput            = 0x0001
	enableLineInput                 = 0x0002
	enableEchoInput                 = 0x0004
	enableProcessedOutput           = 0x0001
	enableVirtualTerminalProcessing = 0x0004
	enableVirtualTerminalInput      = 0x0200
)

var (
	kernel32           = syscall.NewLazyDLL("kernel32.dll")
	procGetStdHandle   = kernel32.NewProc("GetStdHandle")
	procGetConsoleMode = kernel32.NewProc("GetConsoleMode")
	procSetConsoleMode = kernel32.NewProc("SetConsoleMode")

	initOutputOnce sync.Once
	initInputOnce  sync.Once

	inputModeSaved bool
	savedInputMode uint32
)

// getConsoleMode returns the mode of the given standard handle.
func getConsoleMode(handle uintptr) (uint32, bool) {
	var mode uint32
	ret, _, _ := procGetConsoleMode.Call(handle, uintptr(unsafe.Pointer(&mode)))
	if ret == 0 {
		return 0, false
	}
	return mode, true
}

// setConsoleMode sets the mode of the given standard handle.
func setConsoleMode(handle, mode uintptr) bool {
	ret, _, _ := procSetConsoleMode.Call(handle, mode)
	return ret != 0
}

// stdHandle returns one of the standard device handles.
func stdHandle(which uintptr) uintptr {
	handle, _, _ := procGetStdHandle.Call(which)
	return handle
}

// InitOutput enables virtual terminal processing on Windows so that ANSI escape
// sequences are interpreted by the console.
func InitOutput() {
	initOutputOnce.Do(func() {
		handle := stdHandle(stdoutHandle)
		if mode, ok := getConsoleMode(handle); ok {
			setConsoleMode(handle, uintptr(mode|enableProcessedOutput|enableVirtualTerminalProcessing))
		}
	})
}

// initInput switches the console input into a mode that delivers single key
// presses: line input and echo are disabled and virtual terminal input is
// enabled, so that cursor keys arrive as ANSI escape sequences (the same
// behavior as msvcrt.getch in the original implementation).
func initInput() {
	initInputOnce.Do(func() {
		handle := stdHandle(stdinHandle)
		mode, ok := getConsoleMode(handle)
		if !ok {
			return
		}
		inputModeSaved = true
		savedInputMode = mode
		raw := mode &^ uint32(enableProcessedInput|enableLineInput|enableEchoInput)
		raw |= uint32(enableVirtualTerminalInput)
		setConsoleMode(handle, uintptr(raw))
	})
}

// RestoreInput restores the console input mode saved by initInput.
//
// It must be called after the interactive prompt is done, otherwise the parent
// shell keeps running without line input and echo.
func RestoreInput() {
	if !inputModeSaved {
		return
	}
	setConsoleMode(stdHandle(stdinHandle), uintptr(savedInputMode))
	inputModeSaved = false
	initInputOnce = sync.Once{}
}
