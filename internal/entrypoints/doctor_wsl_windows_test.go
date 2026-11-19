//go:build windows

package entrypoints

import (
	"strings"
	"testing"
)

// stubDoctorAncestorNames replaces the ancestor lookup for one test.
func stubDoctorAncestorNames(t *testing.T, names []string) {
	t.Helper()
	original := doctorAncestorNames
	doctorAncestorNames = func() []string { return names }
	t.Cleanup(func() { doctorAncestorNames = original })
}

func TestDoctorWSLInteropCheck(t *testing.T) {
	stubDoctorAncestorNames(t, []string{"powershell.exe", "wslhost.exe", "explorer.exe"})
	check, ok := doctorWSLInteropCheck()
	if !ok {
		t.Fatal("doctorWSLInteropCheck() = false, want true for a WSL host ancestor")
	}
	if check.Status != doctorWarn || check.Name != "wsl" {
		t.Fatalf("check = %+v, want a wsl warning", check)
	}
	if !strings.Contains(check.Message, "inside WSL") {
		t.Fatalf("message = %q", check.Message)
	}
}

func TestDoctorWSLInteropCheckWithoutWSL(t *testing.T) {
	stubDoctorAncestorNames(t, []string{"powershell.exe", "explorer.exe"})
	if check, ok := doctorWSLInteropCheck(); ok {
		t.Fatalf("doctorWSLInteropCheck() = %+v, want no check for a native Windows shell", check)
	}
}
