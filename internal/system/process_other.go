//go:build !windows && !linux

package system

import (
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// psField runs "ps -o <field>= -p <pid>" and returns the trimmed output.
func psField(field string, pid int) string {
	out, err := exec.Command("ps", "-o", field+"=", "-p", strconv.Itoa(pid)).Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// AncestorNames returns the executable names of the calling process ancestors,
// nearest first.
func AncestorNames() []string {
	names := []string{}
	pid := os.Getppid()
	for depth := 0; depth < 10 && pid > 0; depth++ {
		name := psField("comm", pid)
		if name == "" {
			break
		}
		names = append(names, name)
		parent, err := strconv.Atoi(psField("ppid", pid))
		if err != nil || parent <= 0 {
			break
		}
		pid = parent
	}
	return names
}
