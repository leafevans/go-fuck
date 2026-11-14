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
// appends the arguments and the TF_HISTORY value to GOFUCK_ARGS_FILE.
func writeGoFuckStub(t *testing.T, dir string) {
	t.Helper()
	var (
		path    string
		content string
	)
	if runtime.GOOS == "windows" {
		path = filepath.Join(dir, "go-fuck.cmd")
		content = "@echo off\r\n" +
			">> \"%GOFUCK_ARGS_FILE%\" echo ARGS:%*\r\n" +
			">> \"%GOFUCK_ARGS_FILE%\" echo HISTORY:%TF_HISTORY%\r\n" +
			"if defined GOFUCK_STUB_OUTPUT type \"%GOFUCK_STUB_OUTPUT%\"\r\n"
	} else {
		path = filepath.Join(dir, "go-fuck")
		content = "#!/bin/sh\n" +
			"printf 'ARGS:%s\\n' \"$*\" >> \"$GOFUCK_ARGS_FILE\"\n" +
			"printf 'HISTORY:%s\\n' \"${TF_HISTORY-}\" >> \"$GOFUCK_ARGS_FILE\"\n" +
			"if [ -n \"${GOFUCK_STUB_OUTPUT-}\" ]; then cat \"$GOFUCK_STUB_OUTPUT\"; fi\n"
	}
	if err := os.WriteFile(path, []byte(content), 0o755); err != nil {
		t.Fatalf("write the go-fuck stub: %v", err)
	}
}

// aliasInvocation is one invocation of the fake go-fuck command.
type aliasInvocation struct {
	args    string
	history string
}

// runPowershellAliasRaw sources the generated PowerShell alias, runs commands
// with PowerShell and returns the combined output together with the path of the
// recorded arguments file.
func runPowershellAliasRaw(t *testing.T, prelude, commands string) ([]byte, string) {
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
	return output, argsFile
}

// runPowershellAlias sources the generated PowerShell alias and runs commands
// with PowerShell. It returns the recorded invocations of the fake go-fuck
// command.
func runPowershellAlias(t *testing.T, prelude, commands string) []aliasInvocation {
	t.Helper()

	_, argsFile := runPowershellAliasRaw(t, prelude, commands)
	data, err := os.ReadFile(argsFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		t.Fatalf("read the recorded go-fuck arguments: %v", err)
	}

	recorded := []aliasInvocation{}
	current := aliasInvocation{}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(line, "ARGS:"):
			current.args = strings.TrimPrefix(line, "ARGS:")
		case strings.HasPrefix(line, "HISTORY:"):
			current.history = strings.TrimPrefix(line, "HISTORY:")
			recorded = append(recorded, current)
			current = aliasInvocation{}
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
	if len(recorded) != 1 || strings.TrimSpace(recorded[0].args) != "--version" || recorded[0].history != "" {
		t.Fatalf("go-fuck invocation = %+v, want --version without history", recorded)
	}
}

// TestPowershellAliasUsesHistoryWithoutArguments checks that the regular flow
// passes the previous command through TF_HISTORY when no explicit arguments
// are given.
func TestPowershellAliasUsesHistoryWithoutArguments(t *testing.T) {
	resetShellsState(t)

	recorded := runPowershellAlias(t,
		"function Get-History { param($Count) [pscustomobject]@{ CommandLine = 'git status' } }",
		"fuck")
	if len(recorded) != 1 || strings.TrimSpace(recorded[0].args) != "" || recorded[0].history != "git status" {
		t.Fatalf("go-fuck invocation = %+v, want the history through TF_HISTORY", recorded)
	}
}

// TestPowershellAliasHistoryKeepsQuotes checks that a command with embedded
// quotes reaches go-fuck unchanged. Windows PowerShell 5.1 mangles quoted
// arguments when they are passed on the command line, so the alias has to use
// TF_HISTORY.
func TestPowershellAliasHistoryKeepsQuotes(t *testing.T) {
	resetShellsState(t)

	recorded := runPowershellAlias(t,
		"function Get-History { param($Count) [pscustomobject]@{ CommandLine = ('git commit -m ' + [char]34 + 'hello world' + [char]34) } }",
		"fuck")
	if len(recorded) != 1 || strings.TrimSpace(recorded[0].args) != "" {
		t.Fatalf("go-fuck invocation = %+v, want no command line arguments", recorded)
	}
	quote := string(rune(34))
	want := "git commit -m " + quote + "hello world" + quote
	if recorded[0].history != want {
		t.Fatalf("TF_HISTORY = %q, want %q", recorded[0].history, want)
	}
}

// TestPowershellAliasRunsMultilineCorrection checks that a corrected command
// that spans several lines is captured as one value and executed as a whole.
// PowerShell turns the lines of the captured output into an array, so the alias
// has to join them before it looks at the value.
func TestPowershellAliasRunsMultilineCorrection(t *testing.T) {
	resetShellsState(t)

	correctedPath := filepath.Join(t.TempDir(), "corrected.txt")
	corrected := "Write-Output alias-first\r\nWrite-Output alias-second\r\n"
	if err := os.WriteFile(correctedPath, []byte(corrected), 0o600); err != nil {
		t.Fatalf("write the corrected command: %v", err)
	}
	t.Setenv("GOFUCK_STUB_OUTPUT", correctedPath)

	output, _ := runPowershellAliasRaw(t,
		"function Get-History { param($Count) [pscustomobject]@{ CommandLine = 'git status' } }",
		"fuck")
	if !strings.Contains(string(output), "alias-first") || !strings.Contains(string(output), "alias-second") {
		t.Fatalf("PowerShell output = %q, want both lines of the corrected command", output)
	}
}

// TestPowershellAliasPassesInformationCommands checks that the information
// commands run go-fuck directly instead of being evaluated as a corrected
// command: --list-rules prints a rule list, and evaluating it would run the
// rule names as commands.
func TestPowershellAliasPassesInformationCommands(t *testing.T) {
	resetShellsState(t)

	rulesPath := filepath.Join(t.TempDir(), "rules.txt")
	if err := os.WriteFile(rulesPath, []byte("adb_unknown_command\nnot-a-command-marker\n"), 0o600); err != nil {
		t.Fatalf("write the rule list: %v", err)
	}
	t.Setenv("GOFUCK_STUB_OUTPUT", rulesPath)

	output, argsFile := runPowershellAliasRaw(t,
		"function Get-History { param($Count) $null }",
		"fuck --list-rules")

	if !strings.Contains(string(output), "adb_unknown_command") {
		t.Fatalf("PowerShell output = %q, want the printed rule list", output)
	}
	if strings.Contains(string(output), "CommandNotFoundException") {
		t.Fatalf("PowerShell output = %q, want no evaluated rule names", output)
	}
	if !strings.Contains(string(output), "not-a-command-marker") {
		t.Fatalf("PowerShell output = %q, want the raw rule list output", output)
	}

	data, err := os.ReadFile(argsFile)
	if err != nil {
		t.Fatalf("read the recorded go-fuck arguments: %v", err)
	}
	if !strings.Contains(string(data), "ARGS:--list-rules") {
		t.Fatalf("go-fuck arguments = %q, want --list-rules", data)
	}
}
