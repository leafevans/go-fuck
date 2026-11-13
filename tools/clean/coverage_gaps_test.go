package main

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// TestRunReportsStatError covers the stat error branch of run.
func TestRunReportsStatError(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	code := run([]string{"bad\x00path"}, stdout, stderr)
	if code != 1 {
		t.Fatalf("run() = %d, want 1; stderr = %q", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "clean:") {
		t.Fatalf("stderr = %q, want a clean error", stderr.String())
	}
}

// TestRunReportsRemoveError covers the remove error branch of run.
func TestRunReportsRemoveError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("a Windows read only directory does not prevent removing its files")
	}
	dir := t.TempDir()
	sub := filepath.Join(dir, "sub")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatalf("create the directory: %v", err)
	}
	file := filepath.Join(sub, "artifact")
	if err := os.WriteFile(file, []byte("x"), 0o644); err != nil {
		t.Fatalf("write the file: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chmod(sub, 0o755)
		_ = os.Chmod(dir, 0o755)
	})
	if err := os.Chmod(sub, 0o555); err != nil {
		t.Fatalf("make the directory read only: %v", err)
	}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	code := run([]string{file}, stdout, stderr)
	if code == 0 {
		t.Skip("the read only directory file could be removed (running as root?)")
	}
	if code != 1 {
		t.Fatalf("run() = %d, want 1; stderr = %q", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "clean:") {
		t.Fatalf("stderr = %q, want a clean error", stderr.String())
	}
}
