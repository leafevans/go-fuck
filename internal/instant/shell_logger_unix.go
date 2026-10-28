//go:build linux || darwin

package instant

import (
	"errors"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"unsafe"
)

// windowSize is the terminal size structure used by the window size ioctls.
type windowSize struct {
	Rows   uint16
	Cols   uint16
	Xpixel uint16
	Ypixel uint16
}

// Supported reports whether the shell logger can be started on this platform.
func Supported() bool {
	return true
}

// RunShellLogger runs the user shell under a pseudo terminal and copies
// everything it prints into logPath. It returns the exit code of the shell.
func RunShellLogger(logPath string) (int, error) {
	shell := os.Getenv("SHELL")
	if shell == "" {
		return 1, errors.New("SHELL is not set")
	}

	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0o600)
	if err != nil {
		return 1, err
	}
	defer logFile.Close()

	return runLoggedShell(shell, os.Stdin, os.Stdout, NewRingWriter(logFile))
}

// runLoggedShell spawns shell on a pseudo terminal and mirrors its terminal
// input and output. It is separated from RunShellLogger so that tests can
// supply the input and output streams.
func runLoggedShell(shell string, stdin io.Reader, stdout io.Writer, logWriter io.Writer) (int, error) {
	master, slave, err := openPTY()
	if err != nil {
		return 1, err
	}
	defer master.Close()

	restore := makeRawTerminal(stdin)
	if restore != nil {
		defer restore()
	}

	if size, ok := getWindowSize(os.Stdin.Fd()); ok {
		_ = setWindowSize(master.Fd(), size)
	}

	resize := make(chan os.Signal, 1)
	signal.Notify(resize, syscall.SIGWINCH)
	defer signal.Stop(resize)
	stop := make(chan struct{})
	defer close(stop)
	go func() {
		for {
			select {
			case <-resize:
				if size, ok := getWindowSize(os.Stdin.Fd()); ok {
					_ = setWindowSize(master.Fd(), size)
				}
			case <-stop:
				return
			}
		}
	}()

	command := exec.Command(shell)
	command.Stdin = slave
	command.Stdout = slave
	command.Stderr = slave
	command.SysProcAttr = &syscall.SysProcAttr{Setsid: true, Setctty: true, Ctty: 0}
	if err := command.Start(); err != nil {
		slave.Close()
		return 1, err
	}
	slave.Close()

	go func() {
		_, _ = io.Copy(master, stdin)
	}()

	buffer := make([]byte, 4096)
	for {
		count, readErr := master.Read(buffer)
		if count > 0 {
			if _, err := stdout.Write(buffer[:count]); err != nil {
				break
			}
			if _, err := logWriter.Write(buffer[:count]); err != nil {
				break
			}
		}
		if readErr != nil {
			break
		}
	}

	waitErr := command.Wait()
	code := 1
	if command.ProcessState != nil {
		code = command.ProcessState.ExitCode()
	}
	if waitErr == nil {
		code = 0
	}
	if code < 0 {
		code = 1
	}
	return code, nil
}

// makeRawTerminal puts stdin into raw mode and returns a restore function.
// It returns nil when stdin is not a terminal.
func makeRawTerminal(stdin io.Reader) func() {
	file, ok := stdin.(*os.File)
	if !ok {
		return nil
	}
	state, ok := getTermios(file.Fd())
	if !ok {
		return nil
	}
	raw := makeRaw(state)
	if err := setTermios(file.Fd(), raw); err != nil {
		return nil
	}
	return func() {
		_ = setTermios(file.Fd(), state)
	}
}

// makeRaw returns the raw mode attributes of a terminal.
func makeRaw(state *syscall.Termios) *syscall.Termios {
	raw := *state
	raw.Iflag &^= syscall.IGNBRK | syscall.BRKINT | syscall.PARMRK | syscall.ISTRIP |
		syscall.INLCR | syscall.IGNCR | syscall.ICRNL | syscall.IXON
	raw.Oflag &^= syscall.OPOST
	raw.Lflag &^= syscall.ECHO | syscall.ECHONL | syscall.ICANON | syscall.ISIG | syscall.IEXTEN
	raw.Cflag &^= syscall.CSIZE | syscall.PARENB
	raw.Cflag |= syscall.CS8
	raw.Cc[syscall.VMIN] = 1
	raw.Cc[syscall.VTIME] = 0
	return &raw
}

// getWindowSize reads the terminal size of fd.
func getWindowSize(fd uintptr) (windowSize, bool) {
	size := windowSize{}
	if err := ioctl(fd, ioctlTIOCGWINSZ, unsafe.Pointer(&size)); err != nil {
		return windowSize{}, false
	}
	return size, true
}

// setWindowSize sets the terminal size of fd.
func setWindowSize(fd uintptr, size windowSize) error {
	return ioctl(fd, ioctlTIOCSWINSZ, unsafe.Pointer(&size))
}

// ioctl calls the ioctl syscall.
func ioctl(fd, request uintptr, argument unsafe.Pointer) error {
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, fd, request, uintptr(argument))
	if errno != 0 {
		return errno
	}
	return nil
}
