package shells

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/leafevans/go-fuck/internal/conf"
	"github.com/leafevans/go-fuck/internal/consts"
)

// resetShellsState resets the global state used by the shells tests.
func resetShellsState(t *testing.T) {
	t.Helper()
	conf.Reset()
	t.Cleanup(conf.Reset)
}

func TestSplit(t *testing.T) {
	cases := []struct {
		command  string
		expected []string
	}{
		{"ls -la", []string{"ls", "-la"}},
		{"git commit -m 'hello world'", []string{"git", "commit", "-m", "hello world"}},
		{"git commit -m " + string(rune(34)) + "hello world" + string(rune(34)),
			[]string{"git", "commit", "-m", "hello world"}},
		{"echo hello" + string(rune(92)) + " world", []string{"echo", "hello world"}},
		{"", nil},
	}
	for _, testCase := range cases {
		got, err := Split(testCase.command)
		if err != nil {
			t.Errorf("Split(%q) error = %v", testCase.command, err)
			continue
		}
		if strings.Join(got, "|") != strings.Join(testCase.expected, "|") {
			t.Errorf("Split(%q) = %q, want %q", testCase.command, got, testCase.expected)
		}
	}

	if _, err := Split("echo 'unterminated"); err == nil {
		t.Fatal("Split() should fail on an unterminated quote")
	}
}

func TestQuote(t *testing.T) {
	cases := []struct {
		value    string
		expected string
	}{
		{"", "''"},
		{"ls", "ls"},
		{"ls -la", "'ls -la'"},
		{"it's", "'it'" + string(rune(34)) + "'" + string(rune(34)) + "'s'"},
		{"/usr/bin/ls", "/usr/bin/ls"},
		{"a_b@c%+=:,./-", "a_b@c%+=:,./-"},
	}
	for _, testCase := range cases {
		if got := Quote(testCase.value); got != testCase.expected {
			t.Errorf("Quote(%q) = %q, want %q", testCase.value, got, testCase.expected)
		}
	}
}

func TestGenericSplitCommand(t *testing.T) {
	shell := NewGeneric()
	parts := shell.SplitCommand("echo hello" + string(rune(92)) + " world")
	if len(parts) != 2 || parts[0] != "echo" {
		t.Fatalf("SplitCommand() = %q", parts)
	}

	parts = shell.SplitCommand("git commit -m 'unterminated")
	if len(parts) == 0 {
		t.Fatal("SplitCommand() should fall back to whitespace splitting")
	}
}

func TestAndOr(t *testing.T) {
	resetShellsState(t)
	generic := NewGeneric()
	if got := generic.And("a", "b"); got != "a && b" {
		t.Fatalf("And() = %q", got)
	}
	if got := generic.Or("a", "b"); got != "a || b" {
		t.Fatalf("Or() = %q", got)
	}

	fish := NewFish()
	if got := fish.And("a", "b"); got != "a; and b" {
		t.Fatalf("Fish And() = %q", got)
	}
	if got := fish.Or("a", "b"); got != "a; or b" {
		t.Fatalf("Fish Or() = %q", got)
	}

	powershell := NewPowershell()
	if got := powershell.And("a", "b"); got != "(a) -and (b)" {
		t.Fatalf("PowerShell And() = %q", got)
	}
}

func TestAppAlias(t *testing.T) {
	resetShellsState(t)

	aliases := map[string]Shell{
		"bash":       NewBash(),
		"zsh":        NewZsh(),
		"fish":       NewFish(),
		"tcsh":       NewTcsh(),
		"powershell": NewPowershell(),
	}
	for name, shell := range aliases {
		alias := shell.AppAlias("fuck")
		if !strings.Contains(alias, "fuck") {
			t.Errorf("%s AppAlias() does not contain the alias name: %q", name, alias)
		}
		if !strings.Contains(alias, consts.AppName) {
			t.Errorf("%s AppAlias() does not contain the binary name: %q", name, alias)
		}
	}
	if !strings.Contains(NewBash().AppAlias("fuck"), consts.ArgumentPlaceholder) {
		t.Error("bash AppAlias() does not contain the argument placeholder")
	}
}

func TestAppAliasAlterHistory(t *testing.T) {
	resetShellsState(t)
	conf.Current().AlterHistory = false
	if strings.Contains(NewBash().AppAlias("fuck"), "history -s") {
		t.Fatal("bash AppAlias() should not alter history when the setting is disabled")
	}
	if strings.Contains(NewFish().AppAlias("fuck"), "history delete") {
		t.Fatal("fish AppAlias() should not alter history when the setting is disabled")
	}

	conf.Current().AlterHistory = true
	if !strings.Contains(NewBash().AppAlias("fuck"), "history -s") {
		t.Fatal("bash AppAlias() should alter history when the setting is enabled")
	}
}

func TestDetectFromEnv(t *testing.T) {
	resetShellsState(t)
	for name, expected := range map[string]string{
		"bash": "bash", "zsh": "zsh", "fish": "fish", "tcsh": "tcsh",
		"csh": "tcsh", "pwsh": "powershell", "powershell": "powershell",
	} {
		t.Setenv("TF_SHELL", name)
		if got := Detect().Name(); got != expected {
			t.Errorf("Detect() with TF_SHELL=%q = %q, want %q", name, got, expected)
		}
	}
}

func TestInfoGeneric(t *testing.T) {
	generic := NewGeneric()
	if !strings.Contains(generic.Info(), "Generic Shell") {
		t.Fatalf("Info() = %q", generic.Info())
	}
}

func TestBuiltinCommands(t *testing.T) {
	commands := NewGeneric().BuiltinCommands()
	found := false
	for _, command := range commands {
		if command == "cd" {
			found = true
		}
	}
	if !found {
		t.Fatal("BuiltinCommands() does not contain cd")
	}
}

func TestBashHistory(t *testing.T) {
	resetShellsState(t)
	historyPath := filepath.Join(t.TempDir(), "history")
	content := "ls -la" + newline + "git status" + newline
	if err := os.WriteFile(historyPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HISTFILE", historyPath)

	history := NewBash().GetHistory()
	if len(history) != 2 || history[0] != "ls -la" || history[1] != "git status" {
		t.Fatalf("GetHistory() = %q", history)
	}

	conf.Current().HistoryLimit = 1
	if history := NewBash().GetHistory(); len(history) != 1 || history[0] != "git status" {
		t.Fatalf("GetHistory() with a limit = %q", history)
	}
}

func TestZshHistory(t *testing.T) {
	resetShellsState(t)
	historyPath := filepath.Join(t.TempDir(), "zsh_history")
	content := ": 1650000000:0;ls -la" + newline + ": 1650000001:0;git status" + newline
	if err := os.WriteFile(historyPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HISTFILE", historyPath)

	history := NewZsh().GetHistory()
	if len(history) != 2 || history[0] != "ls -la" || history[1] != "git status" {
		t.Fatalf("GetHistory() = %q", history)
	}
}

func TestFishHistory(t *testing.T) {
	resetShellsState(t)
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	historyDir := filepath.Join(home, ".config", "fish")
	if err := os.MkdirAll(historyDir, 0o755); err != nil {
		t.Fatal(err)
	}
	content := "- cmd: ls -la" + newline + "   when: 1650000000" + newline +
		"- cmd: git status" + newline + "   when: 1650000001" + newline
	if err := os.WriteFile(filepath.Join(historyDir, "fish_history"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	history := NewFish().GetHistory()
	if len(history) != 2 || history[0] != "ls -la" || history[1] != "git status" {
		t.Fatalf("GetHistory() = %q", history)
	}
}

func TestTcshHistory(t *testing.T) {
	resetShellsState(t)
	historyPath := filepath.Join(t.TempDir(), "tcsh_history")
	content := "#+1650000000" + newline + "ls -la" + newline +
		"#+1650000001" + newline + "git status" + newline
	if err := os.WriteFile(historyPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HISTFILE", historyPath)

	history := NewTcsh().GetHistory()
	if len(history) != 2 || history[0] != "ls -la" || history[1] != "git status" {
		t.Fatalf("GetHistory() = %q", history)
	}
}

func TestBashAliases(t *testing.T) {
	resetShellsState(t)
	raw := "alias ll='ls -l'" + newline + "alias gs=" + string(rune(34)) +
		"git status" + string(rune(34)) + newline + "invalid line"
	t.Setenv("TF_SHELL_ALIASES", raw)

	aliases := NewBash().GetAliases()
	if aliases["ll"] != "ls -l" {
		t.Fatalf("GetAliases() = %q", aliases)
	}
	if aliases["gs"] != "git status" {
		t.Fatalf("GetAliases() = %q", aliases)
	}

	shell := NewBash()
	if got := shell.FromShell("ll -la"); got != "ls -l -la" {
		t.Fatalf("FromShell() = %q, want the alias to be expanded", got)
	}
	if got := shell.FromShell("ls -la"); got != "ls -la" {
		t.Fatalf("FromShell() = %q", got)
	}
}

func TestZshAliases(t *testing.T) {
	resetShellsState(t)
	t.Setenv("TF_SHELL_ALIASES", "ll='ls -l'"+newline+"gs='git status'")

	aliases := NewZsh().GetAliases()
	if aliases["ll"] != "ls -l" || aliases["gs"] != "git status" {
		t.Fatalf("GetAliases() = %q", aliases)
	}
}

func TestFishOverriddenAliases(t *testing.T) {
	resetShellsState(t)
	t.Setenv("TF_OVERRIDDEN_ALIASES", "myalias,other")
	overridden := NewFish().overriddenAliases()
	for _, name := range []string{"cd", "grep", "ls", "man", "open", "myalias", "other"} {
		if !overridden[name] {
			t.Errorf("alias %q should be overridden", name)
		}
	}
	if overridden["git"] {
		t.Error("git should not be overridden")
	}
}

func TestPowerShellConfiguration(t *testing.T) {
	resetShellsState(t)
	configuration := NewPowershell().HowToConfigure()
	if configuration == nil {
		t.Fatal("HowToConfigure() returned nil")
	}
	if configuration.Path != "$profile" || configuration.CanConfigureAutomatically {
		t.Fatalf("unexpected configuration: %+v", configuration)
	}
	if !strings.Contains(configuration.Content, "--alias") {
		t.Fatalf("unexpected configuration content: %q", configuration.Content)
	}
}

func TestZshParseHistoryLine(t *testing.T) {
	if got := zshParseHistoryLine(": 123:0;ls -la"); got != "ls -la" {
		t.Fatalf("zshParseHistoryLine() = %q", got)
	}
	if got := zshParseHistoryLine("ls -la"); got != "" {
		t.Fatalf("zshParseHistoryLine() = %q, want an empty string", got)
	}
}

func TestFishParseHistoryLine(t *testing.T) {
	if got := fishParseHistoryLine("- cmd: ls -la"); got != "ls -la" {
		t.Fatalf("fishParseHistoryLine() = %q", got)
	}
	if got := fishParseHistoryLine("   when: 1"); got != "" {
		t.Fatalf("fishParseHistoryLine() = %q, want an empty string", got)
	}
}

func TestTcshParseHistoryLine(t *testing.T) {
	if got := tcshParseHistoryLine("#+123"); got != "" {
		t.Fatalf("tcshParseHistoryLine() = %q, want an empty string", got)
	}
	if got := tcshParseHistoryLine("ls -la"); got != "ls -la" {
		t.Fatalf("tcshParseHistoryLine() = %q", got)
	}
}

func TestGenericPutToHistory(t *testing.T) {
	if err := NewGeneric().PutToHistory("ls"); err != nil {
		t.Fatalf("PutToHistory() error = %v", err)
	}
}

func TestFishPutToHistory(t *testing.T) {
	resetShellsState(t)
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	historyDir := filepath.Join(home, ".config", "fish")
	if err := os.MkdirAll(historyDir, 0o755); err != nil {
		t.Fatal(err)
	}
	historyPath := filepath.Join(historyDir, "fish_history")
	if err := os.WriteFile(historyPath, nil, 0o644); err != nil {
		t.Fatal(err)
	}

	if err := NewFish().PutToHistory("ls -la"); err != nil {
		t.Fatalf("PutToHistory() error = %v", err)
	}
	data, err := os.ReadFile(historyPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "- cmd: ls -la") {
		t.Fatalf("unexpected history content: %q", data)
	}
}
