//go:build linux

package system

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

// parentProcessVisible reports whether the parent process that started the test
// binary is still present in /proc. Automation harnesses may launch the test
// through a short lived helper process, so the test can only check the
// ancestor chain while the parent is still alive.
func parentProcessVisible() bool {
	_, err := os.Stat(filepath.Join("/proc", strconv.Itoa(os.Getppid())))
	return err == nil
}

// TestProcessInfoMissingProcess covers the unreadable process branches of
// processInfo.
func TestProcessInfoMissingProcess(t *testing.T) {
	name, parent := processInfo(1 << 30)
	if name != "" || parent != 0 {
		t.Fatalf("processInfo() = %q, %d, want an empty result", name, parent)
	}
}
