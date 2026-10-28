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

// initInput enables virtual terminal input so that cursor keys arrive as ANSI
// escape sequences instead of Windows virtual key codes.
func initInput() {
	initInputOnce.Do(func() {
		handle := stdHandle(stdinHandle)
		if mode, ok := getConsoleMode(handle); ok {
			setConsoleMode(handle, uintptr(mode|enableVirtualTerminalInput))
		}
	})
}
