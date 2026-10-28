//go:build windows

package system

// OpenCommand returns the shell command that opens a file or URL.
func OpenCommand(arg string) string {
	return "cmd /c start " + arg
}
