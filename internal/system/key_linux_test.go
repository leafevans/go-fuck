//go:build linux

package system

import (
	"os"
	"strconv"
	"syscall"
	"testing"
	"unsafe"
)

// Test ioctl constants, see asm-generic/ioctls.h.
const (
	testIoctlTIOCSPTLCK = 0x40045431
	testIoctlTIOCGPTN   = 0x80045430
)

// testIoctl calls the ioctl syscall.
func testIoctl(fd, request uintptr, argument unsafe.Pointer) error {
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, fd, request, uintptr(argument))
	if errno != 0 {
		return errno
	}
	return nil
}

// openTestTerminal opens a pseudo terminal pair for the terminal tests.
func openTestTerminal(t *testing.T) (*os.File, *os.File) {
	t.Helper()
	master, err := os.OpenFile("/dev/ptmx", os.O_RDWR|syscall.O_NOCTTY, 0)
	if err != nil {
		t.Fatalf("open /dev/ptmx: %v", err)
	}
	unlock := int32(0)
	if err := testIoctl(master.Fd(), testIoctlTIOCSPTLCK, unsafe.Pointer(&unlock)); err != nil {
		_ = master.Close()
		t.Fatalf("unlock the pseudo terminal: %v", err)
	}
	var number uint32
	if err := testIoctl(master.Fd(), testIoctlTIOCGPTN, unsafe.Pointer(&number)); err != nil {
		_ = master.Close()
		t.Fatalf("read the pseudo terminal number: %v", err)
	}
	slave, err := os.OpenFile("/dev/pts/"+strconv.FormatUint(uint64(number), 10),
		os.O_RDWR|syscall.O_NOCTTY, 0)
	if err != nil {
		_ = master.Close()
		t.Fatalf("open the pseudo terminal slave: %v", err)
	}
	t.Cleanup(func() {
		_ = master.Close()
		_ = slave.Close()
	})
	return master, slave
}

// TestGetKeyOnTerminal covers the stty success and restore path of GetKey.
func TestGetKeyOnTerminal(t *testing.T) {
	master, slave := openTestTerminal(t)
	previous := os.Stdin
	os.Stdin = slave
	t.Cleanup(func() { os.Stdin = previous })

	if _, err := master.Write([]byte("j")); err != nil {
		t.Fatalf("write the key to the terminal: %v", err)
	}
	key, err := GetKey()
	if err != nil {
		t.Fatalf("GetKey() error = %v", err)
	}
	if string(key) != "j" {
		t.Fatalf("GetKey() = %q, want j", key)
	}
}
