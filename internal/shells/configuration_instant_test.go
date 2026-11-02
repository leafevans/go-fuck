package shells

import (
	"runtime"
	"strings"
	"testing"
)

func TestHowToConfigureInstantContent(t *testing.T) {
	for _, shell := range []Shell{NewBash(), NewZsh()} {
		configuration := shell.HowToConfigure()
		if configuration == nil {
			t.Fatalf("%s HowToConfigure() = nil", shell.Name())
		}
		if !strings.Contains(configuration.InstantContent, "--alias") ||
			!strings.Contains(configuration.InstantContent, "--enable-experimental-instant-mode") {
			t.Fatalf("%s InstantContent = %q, want the instant mode alias line", shell.Name(), configuration.InstantContent)
		}
	}
}

func TestHowToConfigureWithoutInstantMode(t *testing.T) {
	for _, shell := range []Shell{NewFish(), NewTcsh()} {
		configuration := shell.HowToConfigure()
		if configuration == nil {
			t.Fatalf("%s HowToConfigure() = nil", shell.Name())
		}
		if configuration.InstantContent != "" {
			t.Fatalf("%s InstantContent = %q, want an empty string", shell.Name(), configuration.InstantContent)
		}
	}
}

func TestPowerShellHowToConfigureInstantContent(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("PowerShell instant mode is only available on Windows")
	}
	configuration := NewPowershell().HowToConfigure()
	if !strings.Contains(configuration.InstantContent, "--enable-experimental-instant-mode") {
		t.Fatalf("PowerShell InstantContent = %q, want the instant mode flag", configuration.InstantContent)
	}
}
