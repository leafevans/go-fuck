//go:build darwin

package instant

import (
	"os"
	"strings"
	"syscall"
	"unsafe"
)

// Darwin terminal ioctls, see sys/ttycom.h.
const (
	ioctlTIOCGETA     = 0x40487413
	ioctlTIOCSETA     = 0x80487414
	ioctlTIOCGWINSZ   = 0x40087468
	ioctlTIOCSWINSZ   = 0x80087467
	ioctlTIOCPTYGRANT = 0x20007454
	ioctlTIOCPTYUNLK  = 0x20007452
	ioctlTIOCPTYGNAME = 0x40807453
)

// openPTY opens a new pseudo terminal master and slave pair.
func openPTY() (*os.File, *os.File, error) {
	master, err := os.OpenFile("/dev/ptmx", os.O_RDWR|syscall.O_NOCTTY, 0)
	if err != nil {
		return nil, nil, err
	}

	if err := ioctl(master.Fd(), ioctlTIOCPTYGRANT, nil); err != nil {
		master.Close()
		return nil, nil, err
	}
	if err := ioctl(master.Fd(), ioctlTIOCPTYUNLK, nil); err != nil {
		master.Close()
		return nil, nil, err
	}

	var name [128]byte
	if err := ioctl(master.Fd(), ioctlTIOCPTYGNAME, unsafe.Pointer(&name)); err != nil {
		master.Close()
		return nil, nil, err
	}
	slavePath := string(name[:])
	if index := strings.IndexByte(slavePath, 0); index >= 0 {
		slavePath = slavePath[:index]
	}

	slave, err := os.OpenFile(slavePath, os.O_RDWR|syscall.O_NOCTTY, 0)
	if err != nil {
		master.Close()
		return nil, nil, err
	}
	return master, slave, nil
}

// getTermios reads the terminal attributes of fd.
func getTermios(fd uintptr) (*syscall.Termios, bool) {
	state := syscall.Termios{}
	if err := ioctl(fd, ioctlTIOCGETA, unsafe.Pointer(&state)); err != nil {
		return nil, false
	}
	return &state, true
}

// setTermios replaces the terminal attributes of fd.
func setTermios(fd uintptr, state *syscall.Termios) error {
	return ioctl(fd, ioctlTIOCSETA, unsafe.Pointer(state))
}
