package shells

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// writeGoFuckStub writes a fake go-fuck command into dir. Every invocation
// appends its arguments to the file named by GOFUCK_ARGS_FILE.
func writeGoFuckStub(t *testing.T, dir string) {
	t.Helper()

	var (
		path    string
		content string
	)
	if runtime.GOOS == "windows" {
		path = filepath.Join(dir, "go-fuck.cmd")
		content = "@echo off\r\necho %* >> \"%GOFUCK_ARGS_FILE%\"\r\n"
	} else {
		path = filepath.Join(dir, "go-fuck")
		content = "#!/bin/sh\nprintf '%s\\n' \"$*\" >> \"$GOFUCK_ARGS_FILE\"\n"
	}
	if err := os.WriteFile(path, []byte(content), 0o755); err != nil {
		t.Fatalf("write the go-fuck stub: %v", err)
	}
}

// runPowershellAlias sources the generated PowerShell alias and runs commands
// with PowerShell. It returns the arguments that the fake go-fuck command was
// called with, one entry per invocation.
func runPowershellAlias(t *testing.T, prelude, commands string) []string {
	t.Helper()

	powershell := aliasSyntaxPowerShellPath()
	if powershell == "" {
		t.Skip("PowerShell is not available")
	}

	dir := t.TempDir()
	writeGoFuckStub(t, dir)

	aliasPath := filepath.Join(dir, "alias.ps1")
	if err := os.WriteFile(aliasPath, []byte(NewPowershell().AppAlias("fuck")), 0o644); err != nil {
		t.Fatalf("write the alias: %v", err)
	}

	argsFile := filepath.Join(dir, "args.txt")
	scriptPath := filepath.Join(dir, "run.ps1")
	content := prelude + "\n. '" + strings.ReplaceAll(aliasPath, "'", "''") + "'\n" + commands + "\n"
	if err := os.WriteFile(scriptPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write the test script: %v", err)
	}

	t.Setenv("GOFUCK_ARGS_FILE", argsFile)
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))

	arguments := []string{"-NoProfile", "-NonInteractive"}
	if runtime.GOOS == "windows" {
		arguments = append(arguments, "-ExecutionPolicy", "Bypass")
	}
	arguments = append(arguments, "-File", scriptPath)

	output, err := exec.Command(powershell, arguments...).CombinedOutput()
	if err != nil {
		t.Fatalf("PowerShell failed to run the alias: %v\n%s", err, output)
	}

	data, err := os.ReadFile(argsFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		t.Fatalf("read the recorded go-fuck arguments: %v", err)
	}

	recorded := []string{}
	for _, line := range strings.Split(string(data), "\n") {
		if line = strings.TrimSpace(line); line != "" {
			recorded = append(recorded, line)
		}
	}
	return recorded
}

// TestPowershellAliasPassesArgumentsWithoutHistory checks that flags such as
// --version work when the alias is the first command of a session, before
// Get-History contains anything.
func TestPowershellAliasPassesArgumentsWithoutHistory(t *testing.T) {
	resetShellsState(t)

	recorded := runPowershellAlias(t,
		"function Get-History { param($Count) $null }",
		"fuck --version")
	if len(recorded) != 1 || recorded[0] != "--version" {
		t.Fatalf("go-fuck arguments = %q, want [--version]", recorded)
	}
}

// TestPowershellAliasUsesHistoryWithoutArguments checks that the regular flow
// still passes the previous command when no explicit arguments are given.
func TestPowershellAliasUsesHistoryWithoutArguments(t *testing.T) {
	resetShellsState(t)

	recorded := runPowershellAlias(t,
		"function Get-History { param($Count) [pscustomobject]@{ CommandLine = 'git status' } }",
		"fuck")
	// cmd.exe keeps the quotes that PowerShell adds around an argument with a
	// space; the shell here only cares that the command reached go-fuck.
	if len(recorded) != 1 || strings.Trim(recorded[0], "\"") != "git status" {
		t.Fatalf("go-fuck arguments = %q, want [git status]", recorded)
	}
}
