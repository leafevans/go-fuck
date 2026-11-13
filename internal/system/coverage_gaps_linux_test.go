//go:build linux

package system

import "testing"

// TestProcessInfoMissingProcess covers the unreadable process branches of
// processInfo.
func TestProcessInfoMissingProcess(t *testing.T) {
	name, parent := processInfo(1 << 30)
	if name != "" || parent != 0 {
		t.Fatalf("processInfo() = %q, %d, want an empty result", name, parent)
	}
}
