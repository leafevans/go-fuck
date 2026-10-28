//go:build !windows

package system

import (
	"os"
	"os/exec"
	"strings"

	"github.com/leafevans/go-fuck/internal/consts"
)

// sttyState stores a saved terminal state.
type sttyState struct {
	saved string
	valid bool
}

// disableEcho switches the terminal to a mode that delivers key presses
// immediately. The previous state is restored by restore.
//
// The stty utility is used instead of raw termios ioctls so that the same code
// works on Linux, macOS and the BSDs without platform specific syscalls.
func disableEcho() sttyState {
	query := exec.Command("stty", "-g")
	query.Stdin = os.Stdin
	saved, err := query.Output()
	if err != nil {
		return sttyState{}
	}
	raw := exec.Command("stty", "raw", "-echo")
	raw.Stdin = os.Stdin
	if err := raw.Run(); err != nil {
		return sttyState{}
	}
	return sttyState{saved: strings.TrimSpace(string(saved)), valid: true}
}

// restore undoes disableEcho.
func (s sttyState) restore() {
	if !s.valid || s.saved == "" {
		return
	}
	cmd := exec.Command("stty", s.saved)
	cmd.Stdin = os.Stdin
	_ = cmd.Run()
}

// GetKey waits for and returns the next key press.
func GetKey() (consts.Key, error) {
	state := disableEcho()
	defer state.restore()
	return parseKey(stdinByteReader)
}
