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
	procDeleteProcThreadAttributeList     = kernel32DLL.NewProc("DeleteProcThreadAttributeList")
	procUpdateProcThreadAttribute         = kernel32DLL.NewProc("UpdateProcThreadAttribute")
)

// conPTYAvailable reports whether the ConPTY API is available.
func conPTYAvailable() bool {
	procedures := [...]*syscall.LazyProc{
		procCreatePseudoConsole,
		procResizePseudoConsole,
		procClosePseudoConsole,
		procInitializeProcThreadAttributeList,
		procDeleteProcThreadAttributeList,
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
	// #nosec G115 -- signed console coordinates are packed as their two's-complement bit pattern.
	return uintptr(uint32(c.X)&0xffff | (uint32(c.Y)&0xffff)<<16)
}

// hresultFailed reports whether a Windows HRESULT has its failure bit set.
func hresultFailed(result uintptr) bool {
	return result&0x80000000 != 0
}

// hresultHex returns the low 32 bits of an HRESULT for diagnostic output.
func hresultHex(result uintptr) uint64 {
	return uint64(result) & 0xffffffff
}

// conPTY is a pseudo console with a shell process attached to it.
type conPTY struct {
	in, out syscall.Handle
	process syscall.Handle
	// mu protects the pseudoconsole handle through the entire resize call.
	mu  sync.Mutex
	hpc syscall.Handle
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
	// ConPTY owns its copies now. Keeping these pipe ends would hide EOF
	// until the host closed its copies, even after ConPTY had disconnected.
	closeHandles(ptyIn, ptyOut)
	if err != nil {
		closeHandles(cmdIn, cmdOut)
		_, _, _ = procClosePseudoConsole.Call(uintptr(hpc))
		return nil, err
	}
	closeHandles(process.Thread)
	return &conPTY{hpc: hpc, in: cmdIn, out: cmdOut, process: process.Process}, nil
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
	if hresultFailed(result) {
		return 0, fmt.Errorf("CreatePseudoConsole failed: 0x%08x", hresultHex(result))
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

	// The initialized list owns native resources even when updating it or
	// creating the process fails. Keep its backing storage alive through deletion.
	defer func() {
		_, _, _ = procDeleteProcThreadAttributeList.Call(uintptr(unsafe.Pointer(&attributes[0])))
		runtime.KeepAlive(attributes)
	}()

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
func (c *conPTY) wait() shellExit {
	if _, err := syscall.WaitForSingleObject(c.process, waitInfinite); err != nil {
		return shellExit{1, fmt.Errorf("wait for the logged shell: %w", err)}
	}
	var code uint32
	if err := syscall.GetExitCodeProcess(c.process, &code); err != nil {
		return shellExit{1, fmt.Errorf("get the logged shell exit code: %w", err)}
	}
	return shellExit{code: int(code)}
}

// resize resizes the pseudo console.
func (c *conPTY) resize(size coordinate) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	// Hold the lifetime lock through the native call: closeConsole must not
	// invalidate the handle while ResizePseudoConsole is using it.
	hpc := c.hpc
	if hpc == 0 {
		return fmt.Errorf("ResizePseudoConsole failed: the pseudo console is closed")
	}
	result, _, _ := procResizePseudoConsole.Call(uintptr(hpc), size.packed())
	if hresultFailed(result) {
		return fmt.Errorf("ResizePseudoConsole failed: 0x%08x", hresultHex(result))
	}
	return nil
}

// closeConsole shuts down the pseudoconsole. The session coordinator calls
// it while the output loop drains any final frame (unless the read end has
// already been closed).
func (c *conPTY) closeConsole() {
	c.mu.Lock()
	hpc := c.hpc
	c.hpc = 0
	c.mu.Unlock()
	_, _, _ = procClosePseudoConsole.Call(uintptr(hpc))
}

// closeHandles closes the given handles, ignoring invalid ones.
func closeHandles(handles ...syscall.Handle) {
	for _, handle := range handles {
		if handle != 0 {
			_ = syscall.CloseHandle(handle)
		}
	}
}
