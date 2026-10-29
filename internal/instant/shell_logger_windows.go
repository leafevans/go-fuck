//go:build windows

package instant

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"github.com/leafevans/go-fuck/internal/consts"
)

// Windows console mode flags.
const (
	enableProcessedInput            = 0x0001
	enableLineInput                 = 0x0002
	enableEchoInput                 = 0x0004
	enableWindowInput               = 0x0008
	enableMouseInput                = 0x0010
	enableInsertMode                = 0x0020
	enableQuickEditMode             = 0x0040
	enableExtendedFlags             = 0x0080
	enableProcessedOutput           = 0x0001
	enableVirtualTerminalProcessing = 0x0004
	enableVirtualTerminalInput      = 0x0200
)

// defaultColumns and defaultRows are used when the terminal size cannot be
// detected.
const (
	defaultColumns = 120
	defaultRows    = 30
)

// Console procedures used by the shell logger.
var (
	procGetConsoleMode             = kernel32DLL.NewProc("GetConsoleMode")
	procSetConsoleMode             = kernel32DLL.NewProc("SetConsoleMode")
	procGetConsoleScreenBufferInfo = kernel32DLL.NewProc("GetConsoleScreenBufferInfo")
)

// Supported reports whether the shell logger can be started on this platform.
func Supported() bool {
	return conPTYAvailable()
}

// RunShellLogger runs PowerShell under a pseudo console and copies everything
// it prints into logPath. It returns the exit code of the shell.
func RunShellLogger(logPath string) (int, error) {
	if !Supported() {
		return 1, ErrUnsupported
	}
	commandLine, err := shellCommandLine()
	if err != nil {
		return 1, err
	}

	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0o600)
	if err != nil {
		return 1, err
	}
	defer logFile.Close()

	return runLoggedShell(commandLine, os.Stdin, os.Stdout, NewRingWriter(logFile))
}

// shellCommandLine returns the command line of the shell started by the
// logger. The instant mode alias passes the executable of the shell that
// requested the logger, so that the logged shell matches the user's shell.
func shellCommandLine() (string, error) {
	candidates := make([]string, 0, 4)
	if shell := os.Getenv(consts.ShellPathEnv); shell != "" {
		candidates = append(candidates, shell)
	}
	if pwsh, err := exec.LookPath("pwsh.exe"); err == nil {
		candidates = append(candidates, pwsh)
	}
	if powershell, err := exec.LookPath("powershell.exe"); err == nil {
		candidates = append(candidates, powershell)
	}
	if systemRoot := os.Getenv("SystemRoot"); systemRoot != "" {
		candidates = append(candidates, filepath.Join(systemRoot, "System32", "WindowsPowerShell", "v1.0", "powershell.exe"))
	}

	for _, candidate := range candidates {
		if !isPowerShell(candidate) {
			continue
		}
		if _, err := os.Stat(candidate); err != nil {
			continue
		}
		return quoteWindowsCommandLineArg(candidate) + " -NoLogo", nil
	}
	return "", errors.New("PowerShell is not available")
}

// isPowerShell reports whether path points to a PowerShell executable.
func isPowerShell(path string) bool {
	switch strings.ToLower(filepath.Base(path)) {
	case "powershell.exe", "pwsh.exe":
		return true
	}
	return false
}

// quoteWindowsCommandLineArg quotes a file path for a Windows command line.
func quoteWindowsCommandLineArg(argument string) string {
	if argument != "" && !strings.ContainsAny(argument, " 	") {
		return argument
	}
	return string('"') + argument + string('"')
}

// runLoggedShell starts commandLine on a pseudo console and mirrors its
// terminal input and output. It is separated from RunShellLogger so that tests
// can supply the input and output streams.
func runLoggedShell(commandLine string, stdin io.Reader, stdout io.Writer, logWriter io.Writer) (int, error) {
	restore := makeRawConsole(stdin, stdout)
	defer restore()

	size := consoleSize(stdout)
	pty, err := startConPTY(commandLine, size)
	if err != nil {
		return 1, err
	}
	defer pty.close()

	// The output loop and the resize watcher write the log from different
	// goroutines, so the writes are serialized.
	sink := &logSink{writer: logWriter}
	writeLogSize(sink, size)

	exitCodes := make(chan int, 1)
	go func() {
		code := pty.wait()
		// Closing the pseudo console pipes ends the output stream of the
		// output loop below; the process itself has already exited.
		pty.endOutput()
		exitCodes <- code
	}()

	go func() {
		_, _ = io.Copy(pipeWriter{handle: pty.in}, stdin)
	}()

	stopResize := make(chan struct{})
	defer close(stopResize)
	go watchConsoleSize(pty, stdout, stopResize, sink)

	buffer := make([]byte, 4096)
	for {
		var count uint32
		err := syscall.ReadFile(pty.out, buffer, &count, nil)
		if count > 0 {
			if _, writeErr := stdout.Write(buffer[:count]); writeErr != nil {
				pty.closeConsole()
				break
			}
			if _, logErr := sink.Write(buffer[:count]); logErr != nil {
				pty.closeConsole()
				break
			}
		}
		if err != nil || count == 0 {
			break
		}
	}

	return <-exitCodes, nil
}

// watchConsoleSize polls the terminal size and resizes the pseudo console when
// the terminal window is resized. Windows does not signal a resize to console
// applications, so the size has to be polled.
func watchConsoleSize(pty *conPTY, stdout io.Writer, stop <-chan struct{}, sink io.Writer) {
	if _, ok := stdout.(*os.File); !ok {
		return
	}
	current := consoleSize(stdout)
	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			size := consoleSize(stdout)
			if size == current {
				continue
			}
			if err := pty.resize(size); err == nil {
				current = size
				writeLogSize(sink, size)
			}
		}
	}
}

// logSink serializes the writes to the terminal log.
type logSink struct {
	mu     sync.Mutex
	writer io.Writer
}

func (s *logSink) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.writer.Write(p)
}

// writeLogSize records the terminal size in the log. The ConPTY stream encodes
// line breaks and redraws with absolute cursor positions, so the log reader
// needs the terminal size to render it.
func writeLogSize(writer io.Writer, size coordinate) {
	_, _ = fmt.Fprintf(writer, "%s %d %d\n", conPTYSizeRecord, size.X, size.Y)
}

// consoleSize returns the size of the terminal that displays stdout.
func consoleSize(stdout io.Writer) coordinate {
	if file, ok := stdout.(*os.File); ok {
		if size, ok := windowSize(syscall.Handle(file.Fd())); ok {
			return size
		}
	}
	return coordinate{X: defaultColumns, Y: defaultRows}
}

// consoleScreenBufferInfo is the CONSOLE_SCREEN_BUFFER_INFO structure.
type consoleScreenBufferInfo struct {
	Size              coordinate
	CursorPosition    coordinate
	Attributes        uint16
	Window            smallRect
	MaximumWindowSize coordinate
}

// smallRect is the SMALL_RECT structure.
type smallRect struct {
	Left   int16
	Top    int16
	Right  int16
	Bottom int16
}

// windowSize returns the visible size of the console window of handle.
func windowSize(handle syscall.Handle) (coordinate, bool) {
	info := consoleScreenBufferInfo{}
	result, _, _ := procGetConsoleScreenBufferInfo.Call(uintptr(handle), uintptr(unsafe.Pointer(&info)))
	if result == 0 {
		return coordinate{}, false
	}
	columns := int(info.Window.Right) - int(info.Window.Left) + 1
	rows := int(info.Window.Bottom) - int(info.Window.Top) + 1
	if columns <= 0 || rows <= 0 || columns > 0x7FFF || rows > 0x7FFF {
		return coordinate{}, false
	}
	return coordinate{X: int16(columns), Y: int16(rows)}, true
}

// makeRawConsole switches the terminal into the raw mode that the logged shell
// needs and returns a function that restores the previous modes. It does
// nothing when the streams are not consoles.
func makeRawConsole(stdin io.Reader, stdout io.Writer) func() {
	restores := make([]func(), 0, 2)
	if file, ok := stdin.(*os.File); ok {
		handle := syscall.Handle(file.Fd())
		if mode, ok := consoleMode(handle); ok {
			raw := mode &^ uint32(enableProcessedInput|enableLineInput|enableEchoInput|enableWindowInput|enableMouseInput|enableInsertMode|enableQuickEditMode)
			raw |= enableExtendedFlags | enableVirtualTerminalInput
			if setConsoleMode(handle, raw) {
				restores = append(restores, func() { setConsoleMode(handle, mode) })
			}
		}
	}
	if file, ok := stdout.(*os.File); ok {
		handle := syscall.Handle(file.Fd())
		if mode, ok := consoleMode(handle); ok {
			output := mode | enableProcessedOutput | enableVirtualTerminalProcessing
			if setConsoleMode(handle, output) {
				restores = append(restores, func() { setConsoleMode(handle, mode) })
			}
		}
	}
	return func() {
		for index := len(restores) - 1; index >= 0; index-- {
			restores[index]()
		}
	}
}

// consoleMode returns the mode of a console handle.
func consoleMode(handle syscall.Handle) (uint32, bool) {
	var mode uint32
	result, _, _ := procGetConsoleMode.Call(uintptr(handle), uintptr(unsafe.Pointer(&mode)))
	if result == 0 {
		return 0, false
	}
	return mode, true
}

// setConsoleMode replaces the mode of a console handle.
func setConsoleMode(handle syscall.Handle, mode uint32) bool {
	result, _, _ := procSetConsoleMode.Call(uintptr(handle), uintptr(mode))
	return result != 0
}

// pipeWriter writes to the input pipe of the pseudo console.
type pipeWriter struct {
	handle syscall.Handle
}

func (w pipeWriter) Write(p []byte) (int, error) {
	written := 0
	for written < len(p) {
		var count uint32
		if err := syscall.WriteFile(w.handle, p[written:], &count, nil); err != nil {
			return written, err
		}
		if count == 0 {
			return written, io.ErrShortWrite
		}
		written += int(count)
	}
	return written, nil
}
