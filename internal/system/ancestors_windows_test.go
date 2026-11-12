//go:build windows

package system

import "os"

// parentProcessVisible reports whether the parent process that started the test
// binary is still present in the process snapshot. Automation harnesses may
// launch the test through a short lived helper process, so the test can only
// check the ancestor chain while the parent is still alive.
func parentProcessVisible() bool {
	entries := processSnapshot()
	_, ok := entries[uint32(os.Getppid())]
	return ok
}
