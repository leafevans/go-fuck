//go:build linux

package system

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// processInfo returns the executable name and parent pid of a process.
func processInfo(pid int) (string, int) {
	base := filepath.Join("/proc", strconv.Itoa(pid))
	name, err := os.ReadFile(filepath.Join(base, "comm"))
	processName := ""
	if err == nil {
		processName = strings.TrimSpace(string(name))
	}

	stat, err := os.ReadFile(filepath.Join(base, "stat"))
	if err != nil {
		return processName, 0
	}
	// The second field is the command name in parentheses and may contain
	// spaces, so the parent pid is looked up after the last closing paren.
	end := strings.LastIndex(string(stat), ")")
	if end < 0 || end+2 >= len(stat) {
		return processName, 0
	}
	fields := strings.Fields(string(stat[end+2:]))
	if len(fields) < 2 {
		return processName, 0
	}
	parent, err := strconv.Atoi(fields[1])
	if err != nil {
		return processName, 0
	}
	return processName, parent
}

// AncestorNames returns the executable names of the calling process ancestors,
// nearest first.
func AncestorNames() []string {
	names := []string{}
	pid := os.Getppid()
	for depth := 0; depth < 10 && pid > 0; depth++ {
		name, parent := processInfo(pid)
		if name == "" {
			break
		}
		names = append(names, name)
		pid = parent
	}
	return names
}
