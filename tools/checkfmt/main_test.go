package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunFormattedFile(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "formatted.go"), []byte("package example\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	if code := run([]string{dir}, stdout, stderr); code != 0 {
		t.Fatalf("run() = %d, stderr = %q", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "formatted") {
		t.Fatalf("stdout = %q", stdout.String())
	}
}

func TestRunUnformattedFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "unformatted.go")
	if err := os.WriteFile(path, []byte("package example\nfunc  F( ) {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	if code := run([]string{dir}, stdout, stderr); code != 1 {
		t.Fatalf("run() = %d, want 1", code)
	}
	if !strings.Contains(stderr.String(), "unformatted.go") {
		t.Fatalf("stderr = %q, want the unformatted file", stderr.String())
	}
}

func TestRunMissingPath(t *testing.T) {
	stderr := &bytes.Buffer{}
	if code := run([]string{filepath.Join(t.TempDir(), "missing")}, &bytes.Buffer{}, stderr); code != 1 {
		t.Fatalf("run() = %d, want 1", code)
	}
	if stderr.Len() == 0 {
		t.Fatal("run() did not report the gofmt error")
	}
}

func TestRunWithoutGofmt(t *testing.T) {
	t.Setenv("PATH", "")
	stderr := &bytes.Buffer{}
	if code := run(nil, &bytes.Buffer{}, stderr); code != 1 {
		t.Fatalf("run() = %d, want 1", code)
	}
	if !strings.Contains(stderr.String(), "gofmt was not found") {
		t.Fatalf("stderr = %q, want the missing gofmt message", stderr.String())
	}
}
