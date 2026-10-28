// Package system contains the small amount of platform specific code needed by
// go-fuck: executable lookup, terminal initialization, raw key reading and
// process inspection.
package system

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// DevNull is the platform null device path.
const DevNull = os.DevNull

// Which returns the absolute path of an executable, or an empty string when it
// is not available.
func Which(program string) string {
	if program == "" {
		return ""
	}
	path, err := exec.LookPath(program)
	if err != nil {
		return ""
	}
	return path
}

// ExpandHome replaces a leading "~" with the current user home directory.
func ExpandHome(path string) string {
	if path != "~" && !strings.HasPrefix(path, "~"+string(filepath.Separator)) && !strings.HasPrefix(path, "~/") {
		return path
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	if path == "~" {
		return home
	}
	return filepath.Join(home, path[2:])
}
