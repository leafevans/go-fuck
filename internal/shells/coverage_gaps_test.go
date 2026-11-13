package shells

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/leafevans/go-fuck/v2/internal/logs"
)

// TestPowershellChainWithoutCommands covers the empty command list of
// powershellChain.
func TestPowershellChainWithoutCommands(t *testing.T) {
	if got := NewPowershell().And(); got != "" {
		t.Fatalf("And() = %q, want an empty string", got)
	}
	if got := NewPowershell().Or(); got != "" {
		t.Fatalf("Or() = %q, want an empty string", got)
	}
}

// TestQuotePowerShellString covers quotePowerShellString.
func TestQuotePowerShellString(t *testing.T) {
	if got, want := quotePowerShellString("it's"), "'it''s'"; got != want {
		t.Fatalf("quotePowerShellString() = %q, want %q", got, want)
	}
	if got, want := quotePowerShellString("/tmp/go-fuck.log"), "'/tmp/go-fuck.log'"; got != want {
		t.Fatalf("quotePowerShellString() = %q, want %q", got, want)
	}
}

// TestPowershellVersionFallsBackToPwsh covers the pwsh fallback of
// powershellVersion.
func TestPowershellVersionFallsBackToPwsh(t *testing.T) {
	stubShellRunOutput(t, func(name string, args ...string) string {
		if name == "pwsh" {
			return "PowerShell 7.4.1"
		}
		return ""
	})
	if got := powershellVersion(); got != "7.4.1" {
		t.Fatalf("powershellVersion() = %q, want 7.4.1", got)
	}
}

// TestBashProfileConfigurationFallback covers the ~/.bash_profile configuration
// path.
func TestBashProfileConfigurationFallback(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	if err := os.WriteFile(filepath.Join(home, ".bash_profile"), nil, 0o644); err != nil {
		t.Fatalf("write the bash profile: %v", err)
	}

	configuration := NewBash().HowToConfigure()
	if configuration == nil || configuration.Path != "~/.bash_profile" {
		t.Fatalf("HowToConfigure() = %+v, want the bash profile fallback", configuration)
	}
}

// TestFishShellCoverageBranches covers the overridden alias and the unknown
// command branches of the fish shell.
func TestFishShellCoverageBranches(t *testing.T) {
	stubShellRunOutput(t, func(name string, args ...string) string {
		if name != "fish" || len(args) < 2 {
			return ""
		}
		switch args[1] {
		case "functions":
			return ""
		case "alias":
			return "alias ls 'ls --color'\nalias g=git"
		}
		return ""
	})
	fish := NewFish()
	aliases := fish.GetAliases()
	if aliases["g"] != "git" {
		t.Fatalf("GetAliases() = %q, want the g alias", aliases)
	}
	if _, ok := aliases["ls"]; ok {
		t.Fatalf("GetAliases() = %q, want the overridden ls alias to be skipped", aliases)
	}
	if got := fish.FromShell("unknown-command arg"); got != "unknown-command arg" {
		t.Fatalf("FromShell() = %q, want it unchanged", got)
	}
}

// TestFishPutToHistoryFailures covers the missing and unreadable fish history
// file branches of PutToHistory.
func TestFishPutToHistoryFailures(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	if err := NewFish().PutToHistory("ls"); err != nil {
		t.Fatalf("PutToHistory() error = %v, want nil for a missing history file", err)
	}

	historyPath := filepath.Join(home, ".config", "fish", "fish_history")
	if err := os.MkdirAll(historyPath, 0o755); err != nil {
		t.Fatalf("create the history directory: %v", err)
	}
	logs.SetOutput(io.Discard)
	t.Cleanup(func() { logs.SetOutput(os.Stderr) })
	if err := NewFish().PutToHistory("ls"); err != nil {
		t.Fatalf("PutToHistory() error = %v, want nil for an unreadable history file", err)
	}
}

// TestGenericShellHistoryFallbacks covers the shell history without a history
// file.
func TestGenericShellHistoryFallbacks(t *testing.T) {
	if got := NewGeneric().GetHistory(); got != nil {
		t.Fatalf("GetHistory() = %q, want nil for the generic shell", got)
	}

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("HISTFILE", "")
	if got := NewBash().GetHistory(); got != nil {
		t.Fatalf("GetHistory() = %q, want nil for a missing history file", got)
	}
}

// TestZshAliasesSkipMalformedLines covers the malformed alias line branch of
// the zsh shell.
func TestZshAliasesSkipMalformedLines(t *testing.T) {
	t.Setenv("TF_SHELL_ALIASES", "not-an-alias\ng=git")
	aliases := NewZsh().GetAliases()
	if aliases["g"] != "git" {
		t.Fatalf("GetAliases() = %q, want the g alias", aliases)
	}
	if _, ok := aliases["not-an-alias"]; ok {
		t.Fatalf("GetAliases() = %q, want the malformed line to be skipped", aliases)
	}
}

// TestSplitUnterminatedDoubleQuote covers the unterminated double quoted string
// branch of Split.
func TestSplitUnterminatedDoubleQuote(t *testing.T) {
	if _, err := Split("echo \"abc\\"); err != ErrUnterminatedQuote {
		t.Fatalf("Split() error = %v, want ErrUnterminatedQuote", err)
	}
}
