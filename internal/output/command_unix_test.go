//go:build !windows

package output

import (
	"reflect"
	"testing"
)

func TestShellCommandUsesSh(t *testing.T) {
	cmd := shellCommand("echo hello")
	want := []string{"/bin/sh", "-c", "echo hello"}
	if !reflect.DeepEqual(cmd.Args, want) {
		t.Fatalf("shellCommand() args = %q, want %q", cmd.Args, want)
	}
}
