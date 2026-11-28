//go:build windows

package instant

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"github.com/leafevans/go-fuck/v4/internal/consts"
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
	return runShellLogger(logPath, os.Stdin, os.Stdout)
}

// runShellLogger is RunShellLogger with the terminal streams supplied by the
// caller. It refuses to start the logged shell when stdin is not a console:
// non-interactive hosts such as an installer script can load the PowerShell
// profile, and replacing such a process with a logged shell would hang it.
func runShellLogger(logPath string, stdin io.Reader, stdout io.Writer) (int, error) {
	if !Supported() {
		return 1, ErrUnsupported
	}
	if !terminalStdin(stdin) {
		return 1, errors.New("instant mode needs an interactive terminal")
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

	return runLoggedShellContext(context.Background(), commandLine, stdin, stdout, NewRingWriter(logFile), true)
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

// terminalStdin reports whether stdin is an interactive console. Only a console
// reaches end of file when its terminal closes; a pipe or file reaches it for
// other reasons and must not end the logged shell.
func terminalStdin(stdin io.Reader) bool {
	file, ok := stdin.(*os.File)
	if !ok {
		return false
	}
	_, ok = consoleMode(syscall.Handle(file.Fd()))
	return ok
}

// runLoggedShellContext owns the ConPTY session and its child process. It
// borrows the caller's streams. Input EOF ends a console session, but need
// not end a session supplied with a pipe or file.
func runLoggedShellContext(ctx context.Context, commandLine string, stdin io.Reader, stdout io.Writer, logWriter io.Writer, closeOnInputEnd bool) (int, error) {
	ctx, stop := context.WithCancelCause(ctx)
	defer stop(nil)
	if err := ctx.Err(); err != nil {
		return 1, err
	}
	restore := makeRawConsole(stdin, stdout)
	defer restore()

	size := consoleSize(stdout)
	// Validate the initial record before starting a child that needs draining.
	if err := writeLogSize(logWriter, size); err != nil {
		return 1, err
	}
	pty, err := startConPTY(commandLine, size)
	if err != nil {
		return 1, err
	}
	defer closeHandles(pty.process)

	// Output and resize records can be written from different goroutines.
	sink := &logSink{writer: logWriter}
	output := io.MultiWriter(stdout, sink)

	// Transfer the write end to an os.File. A caller-owned Reader can outlive
	// this session if its Read blocks. A late write then sees a closed file,
	// never a reused raw handle.
	inputFile := os.NewFile(uintptr(pty.in), "conpty-input")
	defer inputFile.Close()
	inputDone := make(chan struct{})
	go func() {
		defer close(inputDone)
		_, _ = io.Copy(inputFile, stdin)
	}()
	var terminalDone <-chan struct{}
	if closeOnInputEnd {
		terminalDone = inputDone
	}

	stopResize, resizeDone := make(chan struct{}), make(chan struct{})
	go func() {
		defer close(resizeDone)
		if err := watchConsoleSize(pty, stdout, stopResize, sink); err != nil {
			stop(err)
		}
	}()

	finished := make(chan shellExit, 1)
	go func() {
		exit := waitLoggedShell(ctx, pty, terminalDone)
		// ClosePseudoConsole may block while emitting final output on older
		// Windows. Keep reading until the pipe actually closes: newer Windows
		// can return from ClosePseudoConsole before the final frame arrives.
		pty.closeConsole()
		finished <- exit
	}()

	buffer := make([]byte, 4096)
	for {
		var count uint32
		readErr := syscall.ReadFile(pty.out, buffer, &count, nil)
		if count > 0 && ctx.Err() == nil {
			if _, err := output.Write(buffer[:count]); err != nil {
				stop(err)
			}
		}
		// After a writer failure, keep reading and discard the output. If
		// reading itself failed, close our read end before waiting for ConPTY.
		if readErr != nil || count == 0 {
			closeHandles(pty.out)
			if readErr != nil && readErr != syscall.ERROR_BROKEN_PIPE && readErr != syscall.ERROR_HANDLE_EOF {
				stop(readErr)
			}
			break
		}
	}
	exit := <-finished
	close(stopResize)
	<-resizeDone
	if err := context.Cause(ctx); err != nil {
		return exit.code, err
	}
	return exit.code, exit.err
}

type shellExit struct {
	code int
	err  error
}

// waitLoggedShell preserves the exit status if the shell exits on its own.
// For shutdown, terminate only our child, then wait for it before releasing
// its handles. TerminateProcess can fail if the process has already exited.
func waitLoggedShell(ctx context.Context, pty *conPTY, terminalDone <-chan struct{}) shellExit {
	waited := make(chan shellExit, 1)
	go func() { waited <- pty.wait() }()
	select {
	case exit := <-waited:
		return exit
	case <-ctx.Done():
	case <-terminalDone:
	}
	_ = syscall.TerminateProcess(pty.process, 1)
	return <-waited
}

// watchConsoleSize polls the terminal size and resizes the pseudo console when
// the terminal window is resized. Windows does not signal a resize to console
// applications, so the size has to be polled.
func watchConsoleSize(pty *conPTY, stdout io.Writer, stop <-chan struct{}, sink io.Writer) error {
	if _, ok := stdout.(*os.File); !ok {
		return nil
	}
	current := consoleSize(stdout)
	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-stop:
			return nil
		case <-ticker.C:
			size := consoleSize(stdout)
			if size == current {
				continue
			}
			if err := pty.resize(size); err == nil {
				current = size
				if err := writeLogSize(sink, size); err != nil {
					return err
				}
			}
		}
	}
}

// logSink serializes the writes to the terminal log.
type logSink struct {
	mu     sync.Mutex
	writer io.Writer
}

// Write writes p to the underlying terminal log under the sink mutex.
func (s *logSink) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.writer.Write(p)
}

// writeLogSize records the terminal size in the log. The ConPTY stream encodes
// line breaks and redraws with absolute cursor positions, so the log reader
// needs the terminal size to render it.
func writeLogSize(writer io.Writer, size coordinate) error {
	record := fmt.Sprintf("%s %d %d\n", conPTYSizeRecord, size.X, size.Y)
	n, err := io.WriteString(writer, record)
	if err == nil && n != len(record) {
		return io.ErrShortWrite
	}
	return err
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
		for _, restore := range slices.Backward(restores) {
			restore()
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
