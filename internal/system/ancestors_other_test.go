//go:build !windows && !linux

package system

import "os"

// parentProcessVisible reports whether the parent process that started the test
// binary is still visible to ps. Automation harnesses may launch the test
// through a short lived helper process, so the test can only check the
// ancestor chain while the parent is still alive.
func parentProcessVisible() bool {
	return psField("comm", os.Getppid()) != ""
}
