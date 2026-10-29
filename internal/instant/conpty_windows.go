//go:build windows

package instant

import (
	"fmt"
	"runtime"
	"sync"
	"syscall"
	"unsafe"
)

// Windows constants used by the ConPTY wrapper.
const (
	// extendedStartupInfoPresent tells CreateProcess that the startup
	// information contains a thread attribute list.
	extendedStartupInfoPresent = 0x00080000
	// procThreadAttributePseudoConsole is PROC_THREAD_ATTRIBUTE_PSEUDOCONSOLE,
	// the attribute that attaches a process to a pseudo console.
	procThreadAttributePseudoConsole = 0x00020016
	// waitInfinite is the INFINITE timeout of WaitForSingleObject.
	waitInfinite = 0xFFFFFFFF
)

// kernel32 procedures used by the Windows shell logger. They are resolved
// lazily so that go-fuck keeps working on Windows versions without ConPTY.
var (
	kernel32DLL = syscall.NewLazyDLL("kernel32.dll")

	procCreatePseudoConsole               = kernel32DLL.NewProc("CreatePseudoConsole")
	procResizePseudoConsole               = kernel32DLL.NewProc("ResizePseudoConsole")
	procClosePseudoConsole                = kernel32DLL.NewProc("ClosePseudoConsole")
	procInitializeProcThreadAttributeList = kernel32DLL.NewProc("InitializeProcThreadAttributeList")
	procUpdateProcThreadAttribute         = kernel32DLL.NewProc("UpdateProcThreadAttribute")
)

// conPTYAvailable reports whether the ConPTY API is available.
func conPTYAvailable() bool {
	procedures := [...]*syscall.LazyProc{
		procCreatePseudoConsole,
		procResizePseudoConsole,
		procClosePseudoConsole,
		procInitializeProcThreadAttributeList,
		procUpdateProcThreadAttribute,
	}
	for _, procedure := range procedures {
		if err := procedure.Find(); err != nil {
			return false
		}
	}
	return true
}

// coordinate is the COORD structure of the Windows console API.
type coordinate struct {
	X int16
	Y int16
}

// packed returns the value CreatePseudoConsole and ResizePseudoConsole expect
// for a COORD passed through a register.
func (c coordinate) packed() uintptr {
	return uintptr(uint32(uint16(c.X)) | uint32(uint16(c.Y))<<16)
}

// conPTY is a pseudo console with a shell process attached to it.
type conPTY struct {
	hpc           syscall.Handle
	ptyIn, ptyOut syscall.Handle
	in, out       syscall.Handle
	process       syscall.ProcessInformation
	closeOnce     sync.Once
	// mu guards hpc: the wait goroutine closes the pseudo console while the
	// resize watcher may still be running.
	mu sync.Mutex
}

// startConPTY starts commandLine on a new pseudo console.
func startConPTY(commandLine string, size coordinate) (*conPTY, error) {
	var ptyIn, cmdIn syscall.Handle
	if err := syscall.CreatePipe(&ptyIn, &cmdIn, nil, 0); err != nil {
		return nil, fmt.Errorf("create the input pipe: %w", err)
	}

	var cmdOut, ptyOut syscall.Handle
	if err := syscall.CreatePipe(&cmdOut, &ptyOut, nil, 0); err != nil {
		closeHandles(ptyIn, cmdIn)
		return nil, fmt.Errorf("create the output pipe: %w", err)
	}

	hpc, err := createPseudoConsole(size, ptyIn, ptyOut)
	if err != nil {
		closeHandles(ptyIn, cmdIn, cmdOut, ptyOut)
		return nil, err
	}

	process, err := createProcessAttachedToPseudoConsole(hpc, commandLine)
	if err != nil {
		closeHandles(ptyIn, cmdIn, cmdOut, ptyOut)
		procClosePseudoConsole.Call(uintptr(hpc))
		return nil, err
	}

	return &conPTY{
		hpc:     hpc,
		ptyIn:   ptyIn,
		ptyOut:  ptyOut,
		in:      cmdIn,
		out:     cmdOut,
		process: process,
	}, nil
}

// createPseudoConsole creates a pseudo console that reads its input from input
// and writes its output to output.
func createPseudoConsole(size coordinate, input, output syscall.Handle) (syscall.Handle, error) {
	var hpc syscall.Handle
	result, _, _ := procCreatePseudoConsole.Call(
		size.packed(),
		uintptr(input),
		uintptr(output),
		0,
		uintptr(unsafe.Pointer(&hpc)),
	)
	if int32(result) < 0 {
		return 0, fmt.Errorf("CreatePseudoConsole failed: 0x%08x", uint32(result))
	}
	return hpc, nil
}

// startupInfoEx is STARTUPINFOEXW: the startup information plus the thread
// attribute list that attaches the new process to the pseudo console.
type startupInfoEx struct {
	startupInfo   syscall.StartupInfo
	attributeList unsafe.Pointer
}

// createProcessAttachedToPseudoConsole starts commandLine attached to hpc.
func createProcessAttachedToPseudoConsole(hpc syscall.Handle, commandLine string) (syscall.ProcessInformation, error) {
	var process syscall.ProcessInformation

	commandLinePtr, err := syscall.UTF16PtrFromString(commandLine)
	if err != nil {
		return process, err
	}

	attributeSize := uintptr(0)
	_, _, attributeErr := procInitializeProcThreadAttributeList.Call(0, 1, 0, uintptr(unsafe.Pointer(&attributeSize)))
	if attributeSize == 0 {
		return process, fmt.Errorf("InitializeProcThreadAttributeList failed: %v", attributeErr)
	}

	attributes := make([]byte, attributeSize)
	if result, _, callErr := procInitializeProcThreadAttributeList.Call(
		uintptr(unsafe.Pointer(&attributes[0])),
		1,
		0,
		uintptr(unsafe.Pointer(&attributeSize)),
	); result == 0 {
		return process, fmt.Errorf("InitializeProcThreadAttributeList failed: %w", callErr)
	}

	if result, _, callErr := procUpdateProcThreadAttribute.Call(
		uintptr(unsafe.Pointer(&attributes[0])),
		0,
		procThreadAttributePseudoConsole,
		uintptr(hpc),
		unsafe.Sizeof(hpc),
		0,
		0,
	); result == 0 {
		return process, fmt.Errorf("UpdateProcThreadAttribute failed: %w", callErr)
	}

	startup := startupInfoEx{attributeList: unsafe.Pointer(&attributes[0])}
	startup.startupInfo.Cb = uint32(unsafe.Sizeof(startup))
	startup.startupInfo.Flags = syscall.STARTF_USESTDHANDLES

	err = syscall.CreateProcess(
		nil,
		commandLinePtr,
		nil,
		nil,
		false,
		extendedStartupInfoPresent,
		nil,
		nil,
		&startup.startupInfo,
		&process,
	)
	runtime.KeepAlive(attributes)
	runtime.KeepAlive(commandLinePtr)
	if err != nil {
		return process, fmt.Errorf("CreateProcess failed: %w", err)
	}
	return process, nil
}

// wait blocks until the attached process exits and returns its exit code.
func (c *conPTY) wait() int {
	if _, err := syscall.WaitForSingleObject(c.process.Process, waitInfinite); err != nil {
		return 1
	}
	var code uint32
	if err := syscall.GetExitCodeProcess(c.process.Process, &code); err != nil {
		return 1
	}
	return int(code)
}

// resize resizes the pseudo console.
func (c *conPTY) resize(size coordinate) error {
	c.mu.Lock()
	hpc := c.hpc
	c.mu.Unlock()
	if hpc == 0 {
		return fmt.Errorf("ResizePseudoConsole failed: the pseudo console is closed")
	}
	result, _, _ := procResizePseudoConsole.Call(uintptr(hpc), size.packed())
	if int32(result) < 0 {
		return fmt.Errorf("ResizePseudoConsole failed: 0x%08x", uint32(result))
	}
	return nil
}

// closeConsole closes the pseudo console. Closing it terminates the attached
// process and ends the output stream, so it is safe to call from the goroutine
// that waits for the shell.
func (c *conPTY) closeConsole() {
	c.closeOnce.Do(func() {
		c.mu.Lock()
		hpc := c.hpc
		c.hpc = 0
		c.mu.Unlock()
		if hpc != 0 {
			procClosePseudoConsole.Call(uintptr(hpc))
		}
	})
}

// endOutput closes the pseudo console side handles of the pipes.
//
// It is called after the shell exits: closing the write end of the output pipe
// unblocks the read loop, which still receives the remaining buffered output
// before the pipe reports end of file.
func (c *conPTY) endOutput() {
	c.closeConsole()
	closeHandles(c.ptyIn, c.ptyOut)
	c.ptyIn, c.ptyOut = 0, 0
}

// close releases all handles of the pseudo console.
func (c *conPTY) close() {
	c.closeConsole()
	closeHandles(c.ptyIn, c.ptyOut, c.in, c.out, c.process.Process, c.process.Thread)
	c.ptyIn, c.ptyOut, c.in, c.out = 0, 0, 0, 0
	c.process.Process, c.process.Thread = 0, 0
}

// closeHandles closes the given handles, ignoring invalid ones.
func closeHandles(handles ...syscall.Handle) {
	for _, handle := range handles {
		if handle != 0 {
			_ = syscall.CloseHandle(handle)
		}
	}
}
