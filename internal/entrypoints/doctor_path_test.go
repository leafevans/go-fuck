package entrypoints

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// doctorTestBinaryName returns the executable file name used by the tests.
func doctorTestBinaryName() string {
	if runtime.GOOS == "windows" {
		return "go-fuck.exe"
	}
	return "go-fuck"
}

// writeFakeBinary creates an executable file and returns its path.
func writeFakeBinary(t *testing.T, directory string) string {
	t.Helper()
	path := filepath.Join(directory, doctorTestBinaryName())
	if err := os.MkdirAll(directory, 0o755); err != nil {
		t.Fatalf("os.MkdirAll(%q) error = %v", directory, err)
	}
	if err := os.WriteFile(path, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatalf("os.WriteFile(%q) error = %v", path, err)
	}
	return path
}

// stubDoctorExecutable replaces the executable lookup for one test.
func stubDoctorExecutable(t *testing.T, path string) {
	t.Helper()
	original := doctorExecutable
	doctorExecutable = func() (string, error) { return path, nil }
	t.Cleanup(func() { doctorExecutable = original })
}

// setDoctorPath replaces PATH for one test.
func setDoctorPath(t *testing.T, directories ...string) {
	t.Helper()
	t.Setenv("PATH", strings.Join(directories, string(os.PathListSeparator)))
}

func TestDoctorPathCheckResolvesToThisBuild(t *testing.T) {
	resetEntrypointsState(t)
	dir := t.TempDir()
	running := writeFakeBinary(t, dir)
	stubDoctorExecutable(t, running)
	setDoctorPath(t, dir)

	check := doctorPathCheck(doctorReport{Binary: running})
	if check.Status != doctorOK {
		t.Fatalf("status = %q, want %q; check = %+v", check.Status, doctorOK, check)
	}
	if !strings.Contains(check.Message, "resolves to this build") {
		t.Fatalf("message = %q", check.Message)
	}
}

func TestDoctorPathCheckWarnsAboutEarlierCopy(t *testing.T) {
	resetEntrypointsState(t)
	newDir := t.TempDir()
	oldDir := t.TempDir()
	running := writeFakeBinary(t, newDir)
	oldBinary := writeFakeBinary(t, oldDir)
	stubDoctorExecutable(t, running)
	setDoctorPath(t, oldDir, newDir)

	check := doctorPathCheck(doctorReport{Binary: running})
	if check.Status != doctorWarn {
		t.Fatalf("status = %q, want %q; check = %+v", check.Status, doctorWarn, check)
	}
	if !strings.Contains(check.Message, oldBinary) {
		t.Fatalf("message = %q, want the earlier binary %q", check.Message, oldBinary)
	}
	if !strings.Contains(check.Hint, newDir) {
		t.Fatalf("hint = %q, want the directory of the running binary %q", check.Hint, newDir)
	}
}

func TestDoctorPathCheckWarnsWhenNotOnPath(t *testing.T) {
	resetEntrypointsState(t)
	runningDir := t.TempDir()
	emptyDir := t.TempDir()
	running := writeFakeBinary(t, runningDir)
	stubDoctorExecutable(t, running)
	setDoctorPath(t, emptyDir)

	check := doctorPathCheck(doctorReport{Binary: running})
	if check.Status != doctorWarn {
		t.Fatalf("status = %q, want %q; check = %+v", check.Status, doctorWarn, check)
	}
	if !strings.Contains(check.Message, "not on PATH") {
		t.Fatalf("message = %q", check.Message)
	}
}

func TestDoctorPathCheckReportsLaterCopies(t *testing.T) {
	resetEntrypointsState(t)
	runningDir := t.TempDir()
	otherDir := t.TempDir()
	running := writeFakeBinary(t, runningDir)
	other := writeFakeBinary(t, otherDir)
	stubDoctorExecutable(t, running)
	setDoctorPath(t, runningDir, otherDir)

	check := doctorPathCheck(doctorReport{Binary: running})
	if check.Status != doctorInfo {
		t.Fatalf("status = %q, want %q; check = %+v", check.Status, doctorInfo, check)
	}
	if !strings.Contains(check.Hint, other) {
		t.Fatalf("hint = %q, want the later copy %q", check.Hint, other)
	}
}
