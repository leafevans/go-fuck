package shells

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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
		case "powershell.exe":
			return "5.1.2.3"
		case "pwsh":
			return "PowerShell 7.4.0"
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
	if got := powershellVersion(); got != "5.1.2.3" {
		t.Fatalf("powershellVersion() = %q", got)
	}

	stubShellRunOutput(t, func(string, ...string) string { return "" })
	if fishVersion() != "" || tcshVersion() != "" || powershellVersion() != "" {
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
