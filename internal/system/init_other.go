//go:build !windows

package system

// InitOutput is a no-op on non-Windows platforms, where terminals already
// understand ANSI escape sequences.
func InitOutput() {}

// RestoreInput is a no-op on non-Windows platforms; the terminal state is
// restored right after reading a single key.
func RestoreInput() {}
