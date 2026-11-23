//go:build !windows

package entrypoints

import (
	"strings"
	"testing"
)

func TestDoctorShellBinaryCheckRejectsPowerShellForUnixBinary(t *testing.T) {
	check, ok := doctorShellBinaryCheck(doctorShellState{Name: "powershell", Info: "PowerShell 7"})
	if !ok || check.Status != doctorWarn || !strings.Contains(check.Hint, "Windows") {
		t.Fatalf("doctorShellBinaryCheck() = %+v, %t", check, ok)
	}
	if _, ok := doctorShellBinaryCheck(doctorShellState{Name: "bash", Info: "Bash"}); ok {
		t.Fatal("doctorShellBinaryCheck() reported a mismatch for the native Unix bash")
	}
}
