//go:build !linux && !darwin

package instant

// Supported reports whether the shell logger can be started on this platform.
func Supported() bool {
	return false
}

// RunShellLogger reports that instant mode is only supported on Linux and macOS.
func RunShellLogger(string) (int, error) {
	return 1, ErrUnsupported
}
