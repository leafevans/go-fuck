//go:build linux

package instant

import (
	"os"
	"strconv"
	"syscall"
	"unsafe"
)

// Linux terminal ioctls, see asm-generic/ioctls.h.
const (
	ioctlTCGETS     = 0x5401
	ioctlTCSETS     = 0x5402
	ioctlTIOCGWINSZ = 0x5413
	ioctlTIOCSWINSZ = 0x5414
	ioctlTIOCSPTLCK = 0x40045431
	ioctlTIOCGPTN   = 0x80045430
)

// openPTY opens a new pseudo terminal master and slave pair.
func openPTY() (*os.File, *os.File, error) {
	master, err := os.OpenFile("/dev/ptmx", os.O_RDWR|syscall.O_NOCTTY, 0)
	if err != nil {
		return nil, nil, err
	}

	unlock := int32(0)
	if err := ioctl(master.Fd(), ioctlTIOCSPTLCK, unsafe.Pointer(&unlock)); err != nil {
		master.Close()
		return nil, nil, err
	}

	var number uint32
	if err := ioctl(master.Fd(), ioctlTIOCGPTN, unsafe.Pointer(&number)); err != nil {
		master.Close()
		return nil, nil, err
	}

	slavePath := "/dev/pts/" + strconv.FormatUint(uint64(number), 10)
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
	if err := ioctl(fd, ioctlTCGETS, unsafe.Pointer(&state)); err != nil {
		return nil, false
	}
	return &state, true
}

// setTermios replaces the terminal attributes of fd.
func setTermios(fd uintptr, state *syscall.Termios) error {
	return ioctl(fd, ioctlTCSETS, unsafe.Pointer(state))
}
