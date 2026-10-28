package shells

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/leafevans/go-fuck/internal/consts"
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

func TestAliasSyntaxBash(t *testing.T) {
	resetShellsState(t)

	bash, err := exec.LookPath("bash")
	if err != nil {
		t.Skip("bash is not available")
	}

	bashAlias := NewBash().AppAlias("fuck")
	aliasSyntaxContains(t, "bash", bashAlias)
	if !strings.Contains(bashAlias, consts.ArgumentPlaceholder) {
		t.Fatalf("bash AppAlias() does not contain the argument placeholder: %q", bashAlias)
	}
	aliasSyntaxCommand(t, bash, "-n", "-c", bashAlias)

	genericAlias := NewGeneric().AppAlias("fuck")
	aliasSyntaxContains(t, "generic", genericAlias)
	aliasSyntaxCommand(t, bash, "-n", "-c", genericAlias)
}

func TestAliasSyntaxZsh(t *testing.T) {
	resetShellsState(t)

	zsh, err := exec.LookPath("zsh")
	if err != nil {
		t.Skip("zsh is not available")
	}

	alias := NewZsh().AppAlias("fuck")
	aliasSyntaxContains(t, "zsh", alias)
	if !strings.Contains(alias, consts.ArgumentPlaceholder) {
		t.Fatalf("zsh AppAlias() does not contain the argument placeholder: %q", alias)
	}
	aliasSyntaxCommand(t, zsh, "-n", "-c", alias)
}

func TestAliasSyntaxFish(t *testing.T) {
	resetShellsState(t)

	fish, err := exec.LookPath("fish")
	if err != nil {
		t.Skip("fish is not available")
	}

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
