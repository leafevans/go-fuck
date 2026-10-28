//go:build !windows

package system

// InitOutput is a no-op on non-Windows platforms, where terminals already
// understand ANSI escape sequences.
func InitOutput() {}
