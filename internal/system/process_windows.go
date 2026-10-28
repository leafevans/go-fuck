//go:build windows

package system

import (
	"os"
	"syscall"
	"unsafe"
)

// processEntry32 mirrors the PROCESSENTRY32W structure.
type processEntry32 struct {
	Size            uint32
	CntUsage        uint32
	ProcessID       uint32
	DefaultHeapID   uintptr
	ModuleID        uint32
	CntThreads      uint32
	ParentProcessID uint32
	PriClassBase    int32
	Flags           uint32
	ExeFile         [syscall.MAX_PATH]uint16
}

var (
	procProcess32First = kernel32.NewProc("Process32FirstW")
	procProcess32Next  = kernel32.NewProc("Process32NextW")
	procCreateSnapshot = kernel32.NewProc("CreateToolhelp32Snapshot")
)

// processSnapshot maps process ids to their name and parent pid.
func processSnapshot() map[uint32]processEntry32 {
	snapshot, _, _ := procCreateSnapshot.Call(0x00000002, 0) // TH32CS_SNAPPROCESS
	entries := map[uint32]processEntry32{}
	if snapshot == uintptr(syscall.InvalidHandle) {
		return entries
	}
	defer func() {
		_ = syscall.CloseHandle(syscall.Handle(snapshot))
	}()

	var entry processEntry32
	entry.Size = uint32(unsafe.Sizeof(entry))
	ok, _, _ := procProcess32First.Call(snapshot, uintptr(unsafe.Pointer(&entry)))
	for ok != 0 {
		entries[entry.ProcessID] = entry
		entry.Size = uint32(unsafe.Sizeof(entry))
		ok, _, _ = procProcess32Next.Call(snapshot, uintptr(unsafe.Pointer(&entry)))
	}
	return entries
}

// AncestorNames returns the executable names of the calling process ancestors,
// nearest first.
func AncestorNames() []string {
	names := []string{}
	entries := processSnapshot()
	pid := uint32(os.Getppid())
	for depth := 0; depth < 10 && pid != 0; depth++ {
		entry, ok := entries[pid]
		if !ok {
			break
		}
		names = append(names, syscall.UTF16ToString(entry.ExeFile[:]))
		pid = entry.ParentProcessID
	}
	return names
}
