package shells

import (
	"runtime"
	"strings"
	"testing"

	"github.com/leafevans/go-fuck/internal/consts"
)

// TestPowershellInstantModeInitAlias checks the code that starts the logged
// PowerShell session.
func TestPowershellInstantModeInitAlias(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("PowerShell instant mode is only available on Windows")
	}
	isolateInstantMode(t)

	alias := NewPowershell().InstantModeAlias("fuck")
	for _, want := range []string{"GOFUCK_INSTANT_MODE", "GOFUCK_OUTPUT_LOG", "GOFUCK_SHELL", "--shell-logger", "Remove-Item", "exit"} {
		if !strings.Contains(alias, want) {
			t.Fatalf("instant mode init alias does not contain %q: %q", want, alias)
		}
	}
	if strings.Contains(alias, "function fuck") {
		t.Fatalf("initialization code should not define the alias: %q", alias)
	}
}

// TestPowershellInstantModeMarkedAlias checks the alias installed inside the
// logged shell: it defines the alias and wraps the prompt with the marker.
func TestPowershellInstantModeMarkedAlias(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("PowerShell instant mode is only available on Windows")
	}
	isolateInstantMode(t)
	t.Setenv(consts.InstantModeEnv, "true")

	alias := NewPowershell().InstantModeAlias("fuck")
	for _, want := range []string{"[char]0x200B", "global:prompt", "$function:prompt", "function fuck"} {
		if !strings.Contains(alias, want) {
			t.Fatalf("logged shell alias does not contain %q: %q", want, alias)
		}
	}
	if strings.Contains(alias, consts.UserCommandMark) {
		t.Fatal("generated alias must not embed the marker characters, because the console encoding is not necessarily UTF-8")
	}
}

// TestPowershellInstantModeUnsupportedPlatform checks the fallback on platforms
// without the Windows shell logger.
func TestPowershellInstantModeUnsupportedPlatform(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("instant mode is supported on Windows")
	}
	isolateInstantMode(t)

	alias := NewPowershell().InstantModeAlias("fuck")
	if !strings.Contains(alias, "function fuck") {
		t.Fatalf("fallback alias does not define the alias: %q", alias)
	}
	if strings.Contains(alias, "--shell-logger") {
		t.Fatalf("fallback alias should not start the shell logger: %q", alias)
	}
}

// TestPowershellInstantModeLogPathIsUnique checks that every logged shell gets
// a fresh terminal log.
func TestPowershellInstantModeLogPathIsUnique(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("PowerShell instant mode is only available on Windows")
	}
	isolateInstantMode(t)

	first := NewPowershell().InstantModeAlias("fuck")
	second := NewPowershell().InstantModeAlias("fuck")
	if first == second {
		t.Fatal("every instant mode session must use a fresh log path")
	}
}

// TestPowerShellStringExpression checks the PowerShell code that builds the
// invisible prompt marker.
func TestPowerShellStringExpression(t *testing.T) {
	if got := powerShellStringExpression("aaa"); got != "([string]([char]0x0061) * 3)" {
		t.Fatalf("powerShellStringExpression() = %q", got)
	}
	if got := powerShellStringExpression("ab"); got != "[string]([char]0x0061) + [string]([char]0x0062)" {
		t.Fatalf("powerShellStringExpression() = %q", got)
	}
	if got := powerShellStringExpression(""); got != "" {
		t.Fatalf("powerShellStringExpression() = %q, want empty", got)
	}
}
