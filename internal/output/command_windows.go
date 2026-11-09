//go:build windows

package output

import (
	"encoding/base64"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"unicode/utf16"

	"github.com/leafevans/go-fuck/internal/consts"
	"github.com/leafevans/go-fuck/internal/system"
)

// powerShellExecutable returns the PowerShell executable that started go-fuck,
// or an empty string when the process tree does not contain one.
func powerShellExecutable(ancestors []string) string {
	for _, name := range ancestors {
		switch strings.ToLower(filepath.Base(name)) {
		case "pwsh.exe":
			return "pwsh.exe"
		case "powershell.exe":
			return "powershell.exe"
		}
	}
	return ""
}

// powerShellName returns the PowerShell executable name of path, or an empty
// string when path does not point at a PowerShell executable.
func powerShellName(path string) string {
	switch strings.ToLower(filepath.Base(path)) {
	case "pwsh.exe":
		return "pwsh.exe"
	case "powershell.exe":
		return "powershell.exe"
	}
	return ""
}

// powerShellAncestor returns the PowerShell executable that started go-fuck,
// or an empty string when go-fuck was not started from PowerShell. It is a
// variable so that tests do not depend on the real process tree.
var powerShellAncestor = func() string {
	return powerShellExecutable(system.AncestorNames())
}

// powerShellFromEnv returns the PowerShell executable that the generated alias
// reported through GOFUCK_SHELL, or an empty string. The environment is more
// reliable than the process tree, which some restricted environments do not
// expose completely.
var powerShellFromEnv = func() string {
	path := os.Getenv(consts.ShellPathEnv)
	if path == "" || powerShellName(path) == "" {
		return ""
	}
	if info, err := os.Stat(path); err != nil || info.IsDir() {
		return ""
	}
	return path
}

// powerShellPath resolves a PowerShell executable. PATH is not always complete
// (some processes are started with a reduced environment), so the well-known
// installation path of Windows PowerShell is used as a fallback.
var powerShellPath = func(executable string) string {
	if path, err := exec.LookPath(executable); err == nil {
		return path
	}
	if executable == "powershell.exe" {
		if root := os.Getenv("SystemRoot"); root != "" {
			candidate := filepath.Join(root, "System32", "WindowsPowerShell", "v1.0", executable)
			if _, err := os.Stat(candidate); err == nil {
				return candidate
			}
		}
	}
	return ""
}

// rerunPowerShell returns the executable that re-runs a command started from
// PowerShell, or an empty string when go-fuck was not started from PowerShell.
func rerunPowerShell() string {
	if path := powerShellFromEnv(); path != "" {
		return path
	}
	if executable := powerShellAncestor(); executable != "" {
		return powerShellPath(executable)
	}
	return ""
}

// encodePowerShellCommand encodes a command for -EncodedCommand: the argument
// is the command as UTF-16LE in base64, which avoids all quoting differences
// between the shell that started go-fuck and the shell that re-runs it.
func encodePowerShellCommand(command string) string {
	units := utf16.Encode([]rune(command))
	encoded := make([]byte, 0, len(units)*2)
	for _, unit := range units {
		encoded = append(encoded, byte(unit), byte(unit>>8))
	}
	return base64.StdEncoding.EncodeToString(encoded)
}

// powerShellReRunScript wraps a command so that PowerShell merges all output
// streams into plain text.
//
// PowerShell serializes its error stream as CLIXML when the stream is
// redirected instead of attached to a console, and CLIXML breaks the message
// text into tagged fragments that the rules cannot match. Merging every stream
// inside the shell and printing the exception message of the error records
// keeps the human readable message and drops the positional details that
// PowerShell adds to formatted errors.
func powerShellReRunScript(command string) string {
	return "$ProgressPreference = 'SilentlyContinue'; try { & { " + command +
		" } *>&1 | ForEach-Object { if ($_ -is [System.Management.Automation.ErrorRecord]) { $_.Exception.Message } else { $_ } } } catch { $_.Exception.Message }"
}

// shellCommand builds a command that is executed by the shell that started
// go-fuck.
//
// PowerShell-native commands (cd, rm, mkdir, ...) and PowerShell syntax are not
// understood by cmd.exe, so re-running such a command with cmd.exe captured an
// unrelated "not recognized" error instead of the real one. Commands started
// from PowerShell are therefore re-run with the same PowerShell flavor; every
// other process keeps the cmd.exe behavior.
func shellCommand(command string) *exec.Cmd {
	if path := rerunPowerShell(); path != "" {
		return exec.Command(path, "-NoProfile", "-NonInteractive",
			"-EncodedCommand", encodePowerShellCommand(powerShellReRunScript(command)))
	}
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
