//go:build !windows

package instant

import "strings"

// logLines splits the terminal log written by the POSIX shell logger into text
// lines. The POSIX pseudo terminal writes newlines as plain text, so the log
// can be split directly.
func logLines(data string) []string {
	return strings.Split(data, "\n")
}
