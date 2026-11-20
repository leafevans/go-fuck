//go:build windows

package system

import (
	"os"
	"testing"
)

func TestGetKeyReadsPipedInput(t *testing.T) {
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() error = %v", err)
	}
	if _, err := writer.Write([]byte("x")); err != nil {
		t.Fatalf("write to the pipe: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close the pipe writer: %v", err)
	}

	previous := os.Stdin
	os.Stdin = reader
	t.Cleanup(func() {
		os.Stdin = previous
		RestoreInput()
		_ = reader.Close()
	})

	key, err := GetKey()
	if err != nil {
		t.Fatalf("GetKey() error = %v", err)
	}
	if string(key) != "x" {
		t.Fatalf("GetKey() = %q, want x", key)
	}
}
