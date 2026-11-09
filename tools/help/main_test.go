package main

import (
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
