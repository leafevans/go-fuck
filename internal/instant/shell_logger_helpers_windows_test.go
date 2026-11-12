//go:build windows

package instant

import (
	"bytes"
	"os"
	"path/filepath"
	"syscall"
	"testing"

	"github.com/leafevans/go-fuck/internal/consts"
)

func TestIsPowerShell(t *testing.T) {
	for _, path := range []string{"powershell.exe", "C:/Windows/pwsh.exe", "POWERSHELL.EXE"} {
		if !isPowerShell(path) {
			t.Fatalf("isPowerShell(%q) = false, want true", path)
		}
	}
	for _, path := range []string{"", "bash", "powershell"} {
		if isPowerShell(path) {
			t.Fatalf("isPowerShell(%q) = true, want false", path)
		}
	}
}

func TestQuoteWindowsCommandLineArg(t *testing.T) {
	if got := quoteWindowsCommandLineArg("C:/tools/pwsh.exe"); got != "C:/tools/pwsh.exe" {
		t.Fatalf("quoteWindowsCommandLineArg() = %q", got)
	}
	if got := quoteWindowsCommandLineArg("C:/Program Files/pwsh.exe"); got != `"C:/Program Files/pwsh.exe"` {
		t.Fatalf("quoteWindowsCommandLineArg() = %q", got)
	}
}

func TestShellCommandLineUsesShellPathEnv(t *testing.T) {
	dir := t.TempDir()
	pwsh := filepath.Join(dir, "pwsh.exe")
	if err := os.WriteFile(pwsh, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv(consts.ShellPathEnv, pwsh)
	got, err := shellCommandLine()
	if err != nil {
		t.Fatalf("shellCommandLine() error = %v", err)
	}
	if want := quoteWindowsCommandLineArg(pwsh) + " -NoLogo"; got != want {
		t.Fatalf("shellCommandLine() = %q, want %q", got, want)
	}
}

func TestConsoleHelpersWithInvalidHandle(t *testing.T) {
	if _, ok := consoleMode(syscall.InvalidHandle); ok {
		t.Fatal("consoleMode(InvalidHandle) succeeded")
	}
	if setConsoleMode(syscall.InvalidHandle, 0) {
		t.Fatal("setConsoleMode(InvalidHandle) succeeded")
	}
	if _, ok := windowSize(syscall.InvalidHandle); ok {
		t.Fatal("windowSize(InvalidHandle) succeeded")
	}
}

func TestMakeRawConsoleIgnoresNonFiles(t *testing.T) {
	restore := makeRawConsole(bytes.NewBuffer(nil), &bytes.Buffer{})
	restore()
}

func TestSupportedIsCallable(t *testing.T) {
	_ = Supported()
}
