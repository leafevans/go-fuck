//go:build windows

package instant

import (
	"bytes"
	"os"
	"path/filepath"
	"syscall"
	"testing"

	"github.com/leafevans/go-fuck/v3/internal/consts"
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

func TestCoordinatePacked(t *testing.T) {
	for _, testCase := range []struct {
		name string
		size coordinate
		want uintptr
	}{
		{"zero", coordinate{}, 0},
		{"positive", coordinate{X: 120, Y: 40}, 0x00280078},
		{"negative x", coordinate{X: -1, Y: 2}, 0x0002ffff},
		{"negative y", coordinate{X: 3, Y: -2}, 0xfffe0003},
		{"minimum", coordinate{X: -32768, Y: -32768}, 0x80008000},
		{"maximum", coordinate{X: 32767, Y: 32767}, 0x7fff7fff},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			if got := testCase.size.packed(); got != testCase.want {
				t.Fatalf("coordinate%v.packed() = %#x, want %#x", testCase.size, got, testCase.want)
			}
		})
	}
}

func TestHresultHelpers(t *testing.T) {
	for _, testCase := range []struct {
		result uintptr
		failed bool
		hex    uint64
	}{
		{0, false, 0},
		{0x7fffffff, false, 0x7fffffff},
		{0x80004005, true, 0x80004005},
		{0x180004005, true, 0x80004005},
	} {
		if got := hresultFailed(testCase.result); got != testCase.failed {
			t.Fatalf("hresultFailed(%#x) = %t, want %t", testCase.result, got, testCase.failed)
		}
		if got := hresultHex(testCase.result); got != testCase.hex {
			t.Fatalf("hresultHex(%#x) = %#x, want %#x", testCase.result, got, testCase.hex)
		}
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
