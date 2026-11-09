package conf

import (
	"path/filepath"
	"testing"
)

func TestSetReplacesCurrent(t *testing.T) {
	t.Cleanup(Reset)
	Reset()
	custom := Default()
	custom.WaitCommand = 7
	Set(custom)
	if got := Current().WaitCommand; got != 7 {
		t.Fatalf("Current().WaitCommand = %d, want 7", got)
	}
}

func TestUserDir(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("APPDATA", dir)
	t.Setenv("XDG_CONFIG_HOME", dir)
	t.Setenv("HOME", dir)
	got := UserDir()
	if filepath.Base(got) != "go-fuck" {
		t.Fatalf("UserDir() = %q, want a directory named go-fuck", got)
	}
}
