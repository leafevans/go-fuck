package rules

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/leafevans/go-fuck/v2/internal/utils"
)

const brewUnknownCommandUnknown = `Error: Unknown command: inst`
const brewUnknownCommandUnknownTwo = `Error: Unknown command: instaa`

func TestRuleBrewUnknownCommand(t *testing.T) {
	resetTestState(t)
	rule := Get("brew_unknown_command")
	if rule == nil {
		t.Fatal("brew_unknown_command rule is not registered")
	}

	originalCommands := brewUnknownCommandBrewCommands
	t.Cleanup(func() { brewUnknownCommandBrewCommands = originalCommands })
	brewUnknownCommandBrewCommands = func() []string {
		return append([]string{}, brewUnknownCommandFallbackCommands...)
	}

	assertMatch(t, rule, "brew inst", brewUnknownCommandUnknown)
	for _, command := range brewUnknownCommandBrewCommands() {
		assertNotMatch(t, rule, "brew "+command, "")
	}

	assertNewCommands(t, rule, "brew inst", brewUnknownCommandUnknown,
		[]string{"brew list", "brew install", "brew uninstall"})

	commands, err := rule.GetNewCommand(testCommand("brew instaa", brewUnknownCommandUnknownTwo))
	if err != nil {
		t.Fatalf("rule %q returned an error: %v", rule.Name, err)
	}
	for _, want := range []string{"brew install", "brew uninstall"} {
		if !utils.Contains(commands, want) {
			t.Errorf("rule %q returned %q, want it to contain %q", rule.Name, commands, want)
		}
	}
}

func TestBrewUnknownCommandLists(t *testing.T) {
	prefix := t.TempDir()
	cmdDir := filepath.Join(prefix, "Homebrew", "Library", "Homebrew", "cmd")
	if err := os.MkdirAll(cmdDir, 0o755); err != nil {
		t.Fatal(err)
	}
	for name, content := range map[string]string{
		"test.rb":   "",
		"other.sh":  "",
		"README.md": "",
	} {
		if err := os.WriteFile(filepath.Join(cmdDir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	tapCmd := filepath.Join(prefix, "Homebrew", "Library", "Taps", "user", "homebrew-tap", "cmd")
	if err := os.MkdirAll(tapCmd, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tapCmd, "brew-tapcmd.rb"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	ignoredCmd := filepath.Join(prefix, "Homebrew", "Library", "Taps", "user", "other", "cmd")
	if err := os.MkdirAll(ignoredCmd, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(ignoredCmd, "brew-ignored.rb"), nil, 0o644); err != nil {
		t.Fatal(err)
	}

	commands, err := brewUnknownCommandList(prefix)
	if err != nil {
		t.Fatalf("brewUnknownCommandList() error = %v", err)
	}
	for _, want := range []string{"test", "other", "tapcmd"} {
		if !utils.Contains(commands, want) {
			t.Fatalf("brewUnknownCommandList() = %q, want %q", commands, want)
		}
	}
	if utils.Contains(commands, "ignored") {
		t.Fatalf("brewUnknownCommandList() = %q, want tap commands from homebrew-* taps only", commands)
	}

	if _, err := brewUnknownCommandDirectoryNames(filepath.Join(prefix, "missing")); err == nil {
		t.Fatal("brewUnknownCommandDirectoryNames() succeeded for a missing directory")
	}
	if _, err := brewUnknownCommandDefaultCommands(filepath.Join(prefix, "missing")); err == nil {
		t.Fatal("brewUnknownCommandDefaultCommands() succeeded for a missing directory")
	}
}
