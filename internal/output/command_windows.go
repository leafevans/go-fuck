//go:build windows

package output

import (
	"os"
	"os/exec"
	"strconv"
)

// shellCommand builds a command that is executed by the system shell.
func shellCommand(command string) *exec.Cmd {
	comspec := os.Getenv("COMSPEC")
	if comspec == "" {
		comspec = "cmd.exe"
	}
	return exec.Command(comspec, "/c", command)
}

// killTree kills the shell and all its children.
func killTree(cmd *exec.Cmd) {
	if cmd.Process == nil {
		return
	}
	_ = exec.Command("taskkill", "/F", "/T", "/PID",
		strconv.Itoa(cmd.Process.Pid)).Run()
	_ = cmd.Process.Kill()
}
