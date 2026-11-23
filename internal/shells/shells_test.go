package shells

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"io"

	"github.com/leafevans/go-fuck/v3/internal/conf"
	"github.com/leafevans/go-fuck/v3/internal/consts"

	"github.com/leafevans/go-fuck/v3/internal/logs"
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

func TestFishSplitCommand(t *testing.T) {
	fish := NewFish()
	for _, tc := range []struct {
		command string
		want    []string
	}{
		{"git log 'master'", []string{"git", "log", "master"}},
		{"git log 'repo dir'", []string{"git", "log", "repo dir"}},
		{"git log 'it\\'s'", []string{"git", "log", "it's"}},
		{`git log 'back\\slash'`, []string{"git", "log", `back\slash`}},
		{`git log "double quoted"`, []string{"git", "log", "double quoted"}},
		{`git log escaped\ space`, []string{"git", "log", "escaped space"}},
	} {
		if got := fish.SplitCommand(tc.command); !reflect.DeepEqual(got, tc.want) {
			t.Errorf("SplitCommand(%q) = %q, want %q", tc.command, got, tc.want)
		}
	}
}

func TestTcshSplitCommand(t *testing.T) {
	tcsh := NewTcsh()
	for _, tc := range []struct {
		command string
		want    []string
	}{
		{"git log master", []string{"git", "log", "master"}},
		{`git log "double quoted"`, []string{"git", "log", "double quoted"}},
		{`git log escaped\ space`, []string{"git", "log", "escaped space"}},
	} {
		if got := tcsh.SplitCommand(tc.command); !reflect.DeepEqual(got, tc.want) {
			t.Errorf("SplitCommand(%q) = %q, want %q", tc.command, got, tc.want)
		}
	}
	for _, value := range []string{
		"master", "repo dir", "it's", `back\slash`, "careful!", "line\nbreak",
	} {
		parts := tcsh.SplitCommand("git log " + tcsh.Quote(value))
		if len(parts) != 3 || parts[2] != value {
			t.Errorf("tcsh round trip for %q = %q", value, parts)
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
	if got := powershell.And("a", "b"); got != "a; if ($?) { b }" {
		t.Fatalf("PowerShell And() = %q", got)
	}
	if got := powershell.Or("a", "b"); got != "a; if (-not $?) { b }" {
		t.Fatalf("PowerShell Or() = %q", got)
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

// stubShellRunOutput replaces the command runner used by the shell helpers.
func stubShellRunOutput(t *testing.T, fn func(string, ...string) string) {
	t.Helper()
	original := runOutput
	runOutput = fn
	t.Cleanup(func() { runOutput = original })
}

// stubAncestorNames replaces the process ancestor lookup.
func stubAncestorNames(t *testing.T, names []string) {
	t.Helper()
	original := processAncestorNames
	processAncestorNames = func() []string { return names }
	t.Cleanup(func() { processAncestorNames = original })
}

func TestTrimExeSuffix(t *testing.T) {
	cases := map[string]string{
		"zsh.exe":   "zsh",
		"zsh.EXE":   "zsh",
		".exe":      ".exe",
		"bash":      "bash",
		"a.exe":     "a",
		"x.EXE":     "x",
		"zsh.exe.c": "zsh.exe.c",
	}
	for input, want := range cases {
		if got := trimExeSuffix(input); got != want {
			t.Fatalf("trimExeSuffix(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestFromEnv(t *testing.T) {
	for name, want := range map[string]string{
		"bash":       "bash",
		"zsh":        "zsh",
		"fish":       "fish",
		"tcsh":       "tcsh",
		"powershell": "powershell",
	} {
		t.Setenv("TF_SHELL", name)
		shell := FromEnv()
		if shell == nil || shell.Name() != want {
			t.Fatalf("FromEnv() with TF_SHELL=%q = %v, want %q", name, shell, want)
		}
	}
	t.Setenv("TF_SHELL", "unknown-shell")
	if shell := FromEnv(); shell != nil {
		t.Fatalf("FromEnv() with an unknown shell = %v, want nil", shell)
	}
}

func TestDetect(t *testing.T) {
	t.Setenv("TF_SHELL", "zsh")
	stubAncestorNames(t, []string{"bash"})
	if shell := Detect(); shell.Name() != "zsh" {
		t.Fatalf("Detect() = %q, want the TF_SHELL value zsh", shell.Name())
	}

	t.Setenv("TF_SHELL", "")
	stubAncestorNames(t, []string{`/usr/bin/powershell.exe`, "bash"})
	if shell := Detect(); shell.Name() != "powershell" {
		t.Fatalf("Detect() = %q, want powershell", shell.Name())
	}

	stubAncestorNames(t, []string{"/usr/bin/unknown"})
	if shell := Detect(); shell.Name() != "generic" {
		t.Fatalf("Detect() = %q, want generic", shell.Name())
	}
}

func TestRunOutput(t *testing.T) {
	original := runOutput
	runOutput = func(name string, args ...string) string {
		return original(name, args...)
	}
	t.Cleanup(func() { runOutput = original })

	if got := runOutput("go", "version"); !strings.Contains(got, "go") {
		t.Fatalf("runOutput(go version) = %q", got)
	}
	if got := runOutput("go-fuck-command-that-does-not-exist"); got != "" {
		t.Fatalf("runOutput(missing) = %q, want empty", got)
	}
}

func TestVersionFunctions(t *testing.T) {
	stubShellRunOutput(t, func(name string, args ...string) string {
		switch name {
		case "bash":
			return "5.2.15"
		case "zsh":
			return "5.9"
		case "fish":
			return "fish, version 3.6.0"
		case "tcsh":
			return "tcsh 6.24.00"
		}
		return ""
	})
	if got := bashVersion(); got != "5.2.15" {
		t.Fatalf("bashVersion() = %q", got)
	}
	if got := zshVersion(); got != "5.9" {
		t.Fatalf("zshVersion() = %q", got)
	}
	if got := fishVersion(); got != "3.6.0" {
		t.Fatalf("fishVersion() = %q", got)
	}
	if got := tcshVersion(); got != "6.24.00" {
		t.Fatalf("tcshVersion() = %q", got)
	}
	stubShellRunOutput(t, func(string, ...string) string { return "" })
	if fishVersion() != "" || tcshVersion() != "" {
		t.Fatal("version functions should return an empty string without output")
	}
}

func TestHistoryFileEnvironment(t *testing.T) {
	t.Setenv("HISTFILE", "/tmp/history")
	if got := bashHistoryFile(); got != "/tmp/history" {
		t.Fatalf("bashHistoryFile() = %q", got)
	}
	if got := zshHistoryFile(); got != "/tmp/history" {
		t.Fatalf("zshHistoryFile() = %q", got)
	}
	if got := tcshHistoryFile(); got != "/tmp/history" {
		t.Fatalf("tcshHistoryFile() = %q", got)
	}

	t.Setenv("HISTFILE", "")
	if got := bashHistoryFile(); got != "~/.bash_history" {
		t.Fatalf("bashHistoryFile() = %q", got)
	}
	if got := zshHistoryFile(); got != "~/.zsh_history" {
		t.Fatalf("zshHistoryFile() = %q", got)
	}
	if got := tcshHistoryFile(); got != "~/.history" {
		t.Fatalf("tcshHistoryFile() = %q", got)
	}
	if got := fishHistoryFile(); !strings.Contains(got, "fish_history") {
		t.Fatalf("fishHistoryFile() = %q", got)
	}
}

func TestParseAliasLine(t *testing.T) {
	tests := []struct {
		line, separator, name, value string
		ok                           bool
	}{
		{"ll=ls -l", "=", "ll", "ls -l", true},
		{"ll='ls -l'", "=", "ll", "ls -l", true},
		{`ll="ls -l"`, "=", "ll", "ls -l", true},
		{"=value", "=", "", "", false},
		{"no-separator", "=", "", "", false},
	}
	for _, testCase := range tests {
		name, value, ok := parseAliasLine(testCase.line, testCase.separator)
		if name != testCase.name || value != testCase.value || ok != testCase.ok {
			t.Fatalf("parseAliasLine(%q) = %q, %q, %t; want %q, %q, %t",
				testCase.line, name, value, ok, testCase.name, testCase.value, testCase.ok)
		}
	}
}

func TestShellAliases(t *testing.T) {
	t.Setenv("TF_SHELL_ALIASES", "ll='ls -l'\ng=git")

	zsh := NewZsh()
	if aliases := zsh.GetAliases(); aliases["ll"] != "ls -l" || aliases["g"] != "git" {
		t.Fatalf("zsh.GetAliases() = %q", aliases)
	}
	if got := zsh.FromShell("ll file"); !strings.Contains(got, "ls -l") {
		t.Fatalf("zsh.FromShell() = %q", got)
	}

	t.Setenv("TF_SHELL_ALIASES", "alias ll='ls -l'\nalias g=git")
	bash := NewBash()
	if aliases := bash.GetAliases(); aliases["ll"] != "ls -l" || aliases["g"] != "git" {
		t.Fatalf("bash.GetAliases() = %q", aliases)
	}
	if got := bash.FromShell("g status"); !strings.Contains(got, "git") {
		t.Fatalf("bash.FromShell() = %q", got)
	}

	generic := NewGeneric()
	if aliases := generic.GetAliases(); aliases != nil {
		t.Fatalf("generic.GetAliases() = %q, want nil", aliases)
	}
	if got := generic.FromShell("git status"); got != "git status" {
		t.Fatalf("generic.FromShell() = %q", got)
	}
}

func TestFishAliases(t *testing.T) {
	stubShellRunOutput(t, func(name string, args ...string) string {
		if name != "fish" || len(args) < 2 {
			return ""
		}
		switch args[1] {
		case "functions":
			return "ll\ncd\nfoo\n"
		case "alias":
			return "alias ll 'ls -l'\nalias g=git\nbadline\n"
		}
		return ""
	})
	fish := NewFish()
	aliases := fish.GetAliases()
	if aliases["ll"] != "'ls -l'" || aliases["g"] != "git" || aliases["foo"] != "foo" || aliases["cd"] != "" {
		t.Fatalf("fish.GetAliases() = %q", aliases)
	}
	if got := fish.FromShell("ll file"); got == "ll file" {
		t.Fatalf("fish.FromShell() did not expand ll: %q", got)
	}
	if got := fish.FromShell("foo bar"); !strings.Contains(got, "fish -ic") {
		t.Fatalf("fish.FromShell(foo) = %q", got)
	}
}

func TestTcshAliases(t *testing.T) {
	stubShellRunOutput(t, func(name string, args ...string) string {
		if name == "tcsh" {
			return "ll\tls -l\ncd\tcd\ninvalid-line"
		}
		return ""
	})
	tcsh := NewTcsh()
	aliases := tcsh.GetAliases()
	if aliases["ll"] != "ls -l" || aliases["cd"] != "cd" {
		t.Fatalf("tcsh.GetAliases() = %q", aliases)
	}
	if got := tcsh.FromShell("ll file"); got != "ls -l file" {
		t.Fatalf("tcsh.FromShell() = %q", got)
	}
}

func TestSafeVersionAndInfo(t *testing.T) {
	if got := safeVersion(func() string { panic("boom") }); got != "" {
		t.Fatalf("safeVersion(panic) = %q, want empty", got)
	}
	if got := safeVersion(func() string { return "1.2.3" }); got != "1.2.3" {
		t.Fatalf("safeVersion() = %q", got)
	}
	stubShellRunOutput(t, func(string, ...string) string { return "5.2" })
	if got := NewBash().Info(); !strings.Contains(got, "Bash") || !strings.Contains(got, "5.2") {
		t.Fatalf("Info() = %q", got)
	}
}

func TestCreateConfiguration(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	if err := os.WriteFile(filepath.Join(home, ".bashrc"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	configuration := NewBash().HowToConfigure()
	if configuration == nil || !configuration.CanConfigureAutomatically {
		t.Fatalf("HowToConfigure() = %+v, want it to detect ~/.bashrc", configuration)
	}

	t.Setenv("HOME", t.TempDir())
	t.Setenv("USERPROFILE", t.TempDir())
	configuration = NewBash().HowToConfigure()
	if configuration == nil || configuration.CanConfigureAutomatically {
		t.Fatalf("HowToConfigure() = %+v, want no automatic configuration", configuration)
	}
}

func TestShellInstantModeAliasesDoNotStartALogger(t *testing.T) {
	t.Setenv("GOFUCK_INSTANT_MODE", "")
	for _, shell := range []Shell{NewFish(), NewTcsh(), NewGeneric()} {
		alias := shell.InstantModeAlias("fuck")
		if !strings.Contains(alias, "fuck") {
			t.Fatalf("%s.InstantModeAlias() = %q", shell.Name(), alias)
		}
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
