package main

import (
	"strings"
	"testing"
)

// TestParseScannerError covers the scanner error branch of parse.
func TestParseScannerError(t *testing.T) {
	path := writeMakefile(t, "build: ## Build the go-fuck binary\n"+strings.Repeat("a", 100_000)+"\n")
	if _, err := parse(path); err == nil {
		t.Fatal("parse() of an overlong line returned no error")
	}
}
