package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeMakefile writes a Makefile with the given content to a temporary file.
func writeMakefile(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "Makefile")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}
	return path
}

func TestParse(t *testing.T) {
	path := writeMakefile(t, strings.Join([]string{
		"## Build the binary",
		"build: ## Build the go-fuck binary",
		"build := not-a-target",
		"test: ## Run the tests",
		"",
	}, "\n"))
	targets, err := parse(path)
	if err != nil {
		t.Fatalf("parse() error = %v", err)
	}
	if len(targets) != 2 {
		t.Fatalf("parse() returned %d targets, want 2", len(targets))
	}
	if targets[0].name != "build" || targets[0].description != "Build the go-fuck binary" {
		t.Fatalf("targets[0] = %+v", targets[0])
	}
	if targets[1].name != "test" {
		t.Fatalf("targets[1] = %+v", targets[1])
	}
}

func TestParseErrors(t *testing.T) {
	if _, err := parse(filepath.Join(t.TempDir(), "missing")); err == nil {
		t.Fatalf("parse() of a missing file returned no error")
	}
	path := writeMakefile(t, "all:\n\t@echo hi\n")
	if _, err := parse(path); err == nil {
		t.Fatalf("parse() of a Makefile without documented targets returned no error")
	}
}

func TestRunPrintsTargets(t *testing.T) {
	path := writeMakefile(t, "build: ## Build the go-fuck binary\ntest: ## Run the tests\n")
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	if code := run([]string{"-makefile", path}, stdout, stderr); code != 0 {
		t.Fatalf("run() = %d, stderr = %q", code, stderr.String())
	}
	output := stdout.String()
	for _, want := range []string{"Usage: make TARGET", "build", "Build the go-fuck binary", "test"} {
		if !strings.Contains(output, want) {
			t.Fatalf("run() output = %q, want it to contain %q", output, want)
		}
	}
}

func TestRunErrors(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	if code := run([]string{"-makefile", filepath.Join(t.TempDir(), "missing")}, stdout, stderr); code != 1 {
		t.Fatalf("run() = %d, want 1", code)
	}
	if !strings.Contains(stderr.String(), "help:") {
		t.Fatalf("stderr = %q, want a help error", stderr.String())
	}

	stderr.Reset()
	if code := run([]string{"-unknown"}, stdout, stderr); code != 2 {
		t.Fatalf("run() = %d, want 2 for an invalid flag", code)
	}
	if code := run([]string{"-h"}, stdout, stderr); code != 0 {
		t.Fatalf("run(-h) = %d, want 0", code)
	}
}
