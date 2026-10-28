//go:build !windows

package system

// OpenCommand returns the shell command that opens a file or URL.
func OpenCommand(arg string) string {
	if Which("xdg-open") != "" {
		return "xdg-open " + arg
	}
	return "open " + arg
}
