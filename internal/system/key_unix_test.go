//go:build !windows

package system

import (
	"os"
	"testing"
)

// TestGetKeyFromPipe covers GetKey and the failure branch of disableEcho when
// standard input is not a terminal.
func TestGetKeyFromPipe(t *testing.T) {
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() error = %v", err)
	}
	t.Cleanup(func() {
		_ = reader.Close()
		_ = writer.Close()
	})
	if _, err := writer.Write([]byte("x")); err != nil {
		t.Fatalf("write to the pipe: %v", err)
	}

	previous := os.Stdin
	os.Stdin = reader
	t.Cleanup(func() { os.Stdin = previous })

	key, err := GetKey()
	if err != nil {
		t.Fatalf("GetKey() error = %v", err)
	}
	if string(key) != "x" {
		t.Fatalf("GetKey() = %q, want x", key)
	}
}
