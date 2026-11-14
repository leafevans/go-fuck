package shells

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/leafevans/go-fuck/v2/internal/consts"
)

// aliasSyntaxContains verifies that the generated alias mentions the alias name
// and the go-fuck binary.
func aliasSyntaxContains(t *testing.T, shellName, alias string) {
	t.Helper()
	if !strings.Contains(alias, "fuck") {
		t.Fatalf("%s AppAlias() does not contain the alias name: %q", shellName, alias)
	}
	if !strings.Contains(alias, consts.AppName) {
		t.Fatalf("%s AppAlias() does not contain the binary name: %q", shellName, alias)
	}
}

// aliasSyntaxCommand runs a syntax check of the alias with the given program.
func aliasSyntaxCommand(t *testing.T, program string, args ...string) {
	t.Helper()
	output, err := exec.Command(program, args...).CombinedOutput()
	if err != nil {
		t.Fatalf("%s failed to parse the alias: %v\n%s", program, err, output)
	}
}

// aliasSyntaxFile checks the syntax of an alias by writing it to a file and
// running the shell parser on the file.
//
// Passing a multi-line alias through the command line is fragile: on Windows the
// bash.exe launcher does not always preserve the argument, and bash then fails
// with a syntax error even though the alias is valid. A file is parsed the same
// way by every shell and every launcher.
func aliasSyntaxFile(t *testing.T, program, alias string) {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "alias"), []byte(alias), 0o644); err != nil {
		t.Fatalf("write the alias to a temporary file: %v", err)
	}
	command := exec.Command(program, "-n", "alias")
	command.Dir = dir
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("%s failed to parse the alias file: %v\n%s", program, err, output)
	}
}

// aliasSyntaxRequireShell skips the test when the shell executable cannot
// actually be started.
//
// On Windows, C:\Windows\System32\bash.exe resolves even when WSL is not
// installed or has no distribution; starting it then fails with
// E_ACCESSDENIED instead of running bash. Such an environment cannot check the
// generated alias syntax, so the test has to be skipped rather than failed.
func aliasSyntaxRequireShell(t *testing.T, program string, args ...string) {
	t.Helper()
	output, err := exec.Command(program, args...).CombinedOutput()
	if err != nil {
		t.Skipf("%s cannot be started: %v\n%s", program, err, output)
	}
}

func TestAliasSyntaxBash(t *testing.T) {
	resetShellsState(t)

	bash, err := exec.LookPath("bash")
	if err != nil {
		t.Skip("bash is not available")
	}
	aliasSyntaxRequireShell(t, bash, "-c", "exit 0")

	bashAlias := NewBash().AppAlias("fuck")
	aliasSyntaxContains(t, "bash", bashAlias)
	if !strings.Contains(bashAlias, consts.ArgumentPlaceholder) {
		t.Fatalf("bash AppAlias() does not contain the argument placeholder: %q", bashAlias)
	}
	aliasSyntaxFile(t, bash, bashAlias)

	genericAlias := NewGeneric().AppAlias("fuck")
	aliasSyntaxContains(t, "generic", genericAlias)
	aliasSyntaxFile(t, bash, genericAlias)
}

func TestAliasSyntaxZsh(t *testing.T) {
	resetShellsState(t)

	zsh, err := exec.LookPath("zsh")
	if err != nil {
		t.Skip("zsh is not available")
	}
	aliasSyntaxRequireShell(t, zsh, "-c", "exit 0")

	alias := NewZsh().AppAlias("fuck")
	aliasSyntaxContains(t, "zsh", alias)
	if !strings.Contains(alias, consts.ArgumentPlaceholder) {
		t.Fatalf("zsh AppAlias() does not contain the argument placeholder: %q", alias)
	}
	aliasSyntaxFile(t, zsh, alias)
}

func TestAliasSyntaxFish(t *testing.T) {
	resetShellsState(t)

	fish, err := exec.LookPath("fish")
	if err != nil {
		t.Skip("fish is not available")
	}
	aliasSyntaxRequireShell(t, fish, "-c", "exit 0")

	help, err := exec.Command(fish, "--help").CombinedOutput()
	if err != nil || !strings.Contains(string(help), "--no-execute") {
		t.Skipf("fish does not support --no-execute: %v %s", err, help)
	}

	alias := NewFish().AppAlias("fuck")
	aliasSyntaxContains(t, "fish", alias)
	if !strings.Contains(alias, consts.ArgumentPlaceholder) {
		t.Fatalf("fish AppAlias() does not contain the argument placeholder: %q", alias)
	}
	aliasSyntaxCommand(t, fish, "--no-execute", "-c", alias)
}

// aliasSyntaxPowerShellPath returns the PowerShell executable, preferring pwsh
// over the Windows PowerShell and falling back to the well-known Windows
// installation path when PATH does not contain it.
func aliasSyntaxPowerShellPath() string {
	for _, name := range []string{"pwsh", "powershell"} {
		if path, err := exec.LookPath(name); err == nil {
			return path
		}
	}
	if runtime.GOOS != "windows" {
		return ""
	}
	root := os.Getenv("SystemRoot")
	if root == "" {
		root = "C:" + string(os.PathSeparator) + "Windows"
	}
	path := filepath.Join(root, "System32", "WindowsPowerShell", "v1.0", "powershell.exe")
	if info, err := os.Stat(path); err == nil && !info.IsDir() {
		return path
	}
	return ""
}

func TestAliasSyntaxPowerShell(t *testing.T) {
	resetShellsState(t)

	powershell := aliasSyntaxPowerShellPath()
	if powershell == "" {
		t.Skip("PowerShell is not available")
	}
	aliasSyntaxRequireShell(t, powershell, "-NoProfile", "-Command", "exit 0")

	alias := NewPowershell().AppAlias("fuck")
	if !strings.Contains(alias, "fuck") {
		t.Fatalf("powershell AppAlias() does not contain the alias name: %q", alias)
	}
	if !strings.Contains(alias, consts.AppName) {
		t.Fatalf("powershell AppAlias() does not contain the binary name: %q", alias)
	}

	path := filepath.Join(t.TempDir(), "alias.ps1")
	if err := os.WriteFile(path, []byte(alias), 0o644); err != nil {
		t.Fatalf("write the alias to a temporary file: %v", err)
	}
	quoted := strings.ReplaceAll(path, "'", "''")
	validate := "[scriptblock]::Create((Get-Content -Raw '" + quoted +
		"')) | Out-Null; exit 0"
	aliasSyntaxCommand(t, powershell, "-NoProfile", "-Command", validate)
}
