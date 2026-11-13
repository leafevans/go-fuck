package entrypoints

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/leafevans/go-fuck/internal/argumentparser"
	"github.com/leafevans/go-fuck/internal/shells"
)

// readFile reads a file in tests.
func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(data)
}

// stubPowerShellProfiles replaces the PowerShell profile resolution for tests.
func stubPowerShellProfiles(t *testing.T, paths []string) {
	t.Helper()
	original := powerShellProfiles
	powerShellProfiles = func() []string { return paths }
	t.Cleanup(func() { powerShellProfiles = original })
}

func TestMainSetupBash(t *testing.T) {
	configDir := resetEntrypointsState(t)
	t.Setenv("TF_SHELL", "bash")
	configPath := filepath.Join(configDir, ".bashrc")
	if err := os.WriteFile(configPath, []byte("export EXISTING=1\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	code, _, stderr := runMain(t, "go-fuck", "--setup")
	if code != 0 {
		t.Fatalf("Main(--setup) = %d, want 0; stderr = %s", code, stderr)
	}
	if !strings.Contains(stderr, "now configured") {
		t.Fatalf("stderr = %q, want a success message", stderr)
	}
	content := readFile(t, configPath)
	if !strings.Contains(content, "export EXISTING=1") {
		t.Fatalf("existing configuration was lost: %q", content)
	}
	if got := strings.Count(content, "go-fuck --alias"); got != 1 {
		t.Fatalf("configuration contains %d alias lines, want 1: %q", got, content)
	}

	code, _, stderr = runMain(t, "go-fuck", "--setup")
	if code != 0 {
		t.Fatalf("second Main(--setup) = %d, want 0", code)
	}
	if !strings.Contains(stderr, "already configured") {
		t.Fatalf("stderr = %q, want an already configured message", stderr)
	}
	if got := strings.Count(readFile(t, configPath), "go-fuck --alias"); got != 1 {
		t.Fatalf("configuration contains %d alias lines after the second run, want 1", got)
	}
}

func TestMainSetupReplacesLegacyInstantModeLine(t *testing.T) {
	configDir := resetEntrypointsState(t)
	t.Setenv("TF_SHELL", "zsh")
	configPath := filepath.Join(configDir, ".zshrc")
	legacy := "eval $(go-fuck --alias --enable-experimental-instant-mode)\n"
	if err := os.WriteFile(configPath, []byte(legacy), 0o644); err != nil {
		t.Fatal(err)
	}

	code, _, stderr := runMain(t, "go-fuck", "--setup", "--enable-experimental-instant-mode")
	if code != 0 {
		t.Fatalf("Main(--setup --enable-experimental-instant-mode) = %d, want 0; stderr = %s", code, stderr)
	}
	content := readFile(t, configPath)
	if !strings.Contains(content, "eval \"$(go-fuck --alias --enable-experimental-instant-mode)\"") {
		t.Fatalf("legacy line was not upgraded to the quoted form: %q", content)
	}
	if got := strings.Count(content, "--enable-experimental-instant-mode"); got != 1 {
		t.Fatalf("configuration contains %d instant mode lines, want 1: %q", got, content)
	}
}

func TestMainSetupAppendsNewline(t *testing.T) {
	configDir := resetEntrypointsState(t)
	t.Setenv("TF_SHELL", "bash")
	configPath := filepath.Join(configDir, ".bashrc")
	if err := os.WriteFile(configPath, []byte("export EXISTING=1"), 0o644); err != nil {
		t.Fatal(err)
	}

	if code, _, stderr := runMain(t, "go-fuck", "--setup"); code != 0 {
		t.Fatalf("Main(--setup) = %d, want 0; stderr = %s", code, stderr)
	}
	content := readFile(t, configPath)
	if !strings.HasPrefix(content, "export EXISTING=1\n") {
		t.Fatalf("configuration was not separated by a newline: %q", content)
	}
}

func TestMainSetupManualPowerShell(t *testing.T) {
	resetEntrypointsState(t)
	t.Setenv("TF_SHELL", "pwsh")
	stubPowerShellProfiles(t, nil)

	code, _, stderr := runMain(t, "go-fuck", "--setup")
	if code != 0 {
		t.Fatalf("Main(--setup) = %d, want 0", code)
	}
	if !strings.Contains(stderr, "Add the following line") || !strings.Contains(stderr, "go-fuck --alias") {
		t.Fatalf("stderr = %q, want manual configuration instructions", stderr)
	}
}

func TestMainSetupPowerShell(t *testing.T) {
	configDir := resetEntrypointsState(t)
	t.Setenv("TF_SHELL", "powershell")

	existingProfile := filepath.Join(configDir, "Documents", "WindowsPowerShell", "Microsoft.PowerShell_profile.ps1")
	if err := os.MkdirAll(filepath.Dir(existingProfile), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(existingProfile, []byte("# existing profile\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	missingProfile := filepath.Join(configDir, "Documents", "PowerShell", "Microsoft.PowerShell_profile.ps1")
	stubPowerShellProfiles(t, []string{existingProfile, missingProfile})

	code, _, stderr := runMain(t, "go-fuck", "--setup")
	if code != 0 {
		t.Fatalf("Main(--setup) = %d, want 0; stderr = %s", code, stderr)
	}
	if got := strings.Count(stderr, "now configured"); got != 2 {
		t.Fatalf("stderr = %q, want two success messages", stderr)
	}
	if !strings.Contains(stderr, existingProfile) || !strings.Contains(stderr, missingProfile) {
		t.Fatalf("stderr = %q, want both profile paths", stderr)
	}

	if content := readFile(t, existingProfile); !strings.Contains(content, "# existing profile") {
		t.Fatalf("existing profile content was lost: %q", content)
	}
	if got := strings.Count(readFile(t, existingProfile), "go-fuck --alias"); got != 1 {
		t.Fatalf("existing profile contains %d alias lines, want 1", got)
	}
	if got := strings.Count(readFile(t, missingProfile), "go-fuck --alias"); got != 1 {
		t.Fatalf("created profile contains %d alias lines, want 1", got)
	}

	code, _, stderr = runMain(t, "go-fuck", "--setup")
	if code != 0 {
		t.Fatalf("second Main(--setup) = %d, want 0", code)
	}
	if !strings.Contains(stderr, "already configured") {
		t.Fatalf("stderr = %q, want an already configured message", stderr)
	}
	if got := strings.Count(readFile(t, existingProfile), "go-fuck --alias"); got != 1 {
		t.Fatalf("existing profile contains %d alias lines after the second run, want 1", got)
	}
}

func TestMainSetupInstantModeBash(t *testing.T) {
	configDir := resetEntrypointsState(t)
	t.Setenv("TF_SHELL", "bash")
	configPath := filepath.Join(configDir, ".bashrc")
	if err := os.WriteFile(configPath, []byte("export EXISTING=1\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	code, _, stderr := runMain(t, "go-fuck", "--setup", "--enable-experimental-instant-mode")
	if code != 0 {
		t.Fatalf("Main(--setup --enable-experimental-instant-mode) = %d, want 0; stderr = %s", code, stderr)
	}
	content := readFile(t, configPath)
	if !strings.Contains(content, "--alias --enable-experimental-instant-mode") {
		t.Fatalf("configuration does not enable instant mode: %q", content)
	}
}

func TestMainSetupInstantModeUnsupportedShell(t *testing.T) {
	configDir := resetEntrypointsState(t)
	t.Setenv("TF_SHELL", "fish")
	configPath := filepath.Join(configDir, ".config", "fish", "config.fish")
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(configPath, []byte("set -g EXISTING 1\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	code, _, stderr := runMain(t, "go-fuck", "--setup", "--enable-experimental-instant-mode")
	if code != 0 {
		t.Fatalf("Main(--setup --enable-experimental-instant-mode) = %d, want 0; stderr = %s", code, stderr)
	}
	if !strings.Contains(stderr, "instant mode is not supported") {
		t.Fatalf("stderr = %q, want an instant mode warning", stderr)
	}
	content := readFile(t, configPath)
	if strings.Contains(content, "--enable-experimental-instant-mode") {
		t.Fatalf("unsupported shell was configured with instant mode: %q", content)
	}
	if !strings.Contains(content, "go-fuck --alias | source") {
		t.Fatalf("regular fish alias was not configured: %q", content)
	}
}

func TestMainSetupInstantModePowerShell(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("PowerShell instant mode is only available on Windows")
	}
	configDir := resetEntrypointsState(t)
	t.Setenv("TF_SHELL", "powershell")
	profile := filepath.Join(configDir, "Documents", "PowerShell", "Microsoft.PowerShell_profile.ps1")
	stubPowerShellProfiles(t, []string{profile})

	code, _, stderr := runMain(t, "go-fuck", "--setup", "--enable-experimental-instant-mode")
	if code != 0 {
		t.Fatalf("Main(--setup --enable-experimental-instant-mode) = %d, want 0; stderr = %s", code, stderr)
	}
	if content := readFile(t, profile); !strings.Contains(content, "--alias --enable-experimental-instant-mode") {
		t.Fatalf("profile does not enable instant mode: %q", content)
	}
}

func TestMainSetupMissingConfigFile(t *testing.T) {
	resetEntrypointsState(t)
	t.Setenv("TF_SHELL", "zsh")

	code, _, stderr := runMain(t, "go-fuck", "--setup")
	if code != 0 {
		t.Fatalf("Main(--setup) = %d, want 0", code)
	}
	if !strings.Contains(stderr, "~/.zshrc") {
		t.Fatalf("stderr = %q, want the zsh configuration path", stderr)
	}
}

func TestResolveProfiles(t *testing.T) {
	t.Run("keeps the flavor order", func(t *testing.T) {
		lookup := func(executable string) (string, bool) {
			switch executable {
			case "powershell.exe":
				return "C:\\Users\\me\\Documents\\WindowsPowerShell\\Microsoft.PowerShell_profile.ps1", true
			case "pwsh":
				return "C:\\Users\\me\\Documents\\PowerShell\\Microsoft.PowerShell_profile.ps1", true
			}
			return "", false
		}
		paths := resolveProfiles([]string{"powershell.exe", "pwsh"}, lookup)
		if len(paths) != 2 {
			t.Fatalf("resolveProfiles() = %v, want two paths", paths)
		}
		if !strings.Contains(paths[0], "WindowsPowerShell") || !strings.Contains(paths[1], "PowerShell") {
			t.Fatalf("resolveProfiles() = %v, want the Windows PowerShell and PowerShell 7 profiles", paths)
		}
	})

	t.Run("drops missing flavors", func(t *testing.T) {
		lookup := func(executable string) (string, bool) {
			if executable == "pwsh" {
				return "", false
			}
			return "profile.ps1", true
		}
		if paths := resolveProfiles([]string{"powershell.exe", "pwsh"}, lookup); len(paths) != 1 {
			t.Fatalf("resolveProfiles() = %v, want one path", paths)
		}
	})

	t.Run("deduplicates paths", func(t *testing.T) {
		lookup := func(string) (string, bool) { return "profile.ps1", true }
		if paths := resolveProfiles([]string{"powershell.exe", "pwsh"}, lookup); len(paths) != 1 {
			t.Fatalf("resolveProfiles() = %v, want one path", paths)
		}
	})
}

func TestPowerShellProfile(t *testing.T) {
	resolved := 0
	for _, executable := range powerShellProfileExecutables {
		path, ok := powerShellProfile(executable)
		if !ok {
			continue
		}
		resolved++
		if !filepath.IsAbs(path) {
			t.Errorf("powerShellProfile(%q) = %q, want an absolute path", executable, path)
		}
		if !strings.HasSuffix(strings.ToLower(path), ".ps1") {
			t.Errorf("powerShellProfile(%q) = %q, want a .ps1 profile path", executable, path)
		}
	}
	if resolved == 0 {
		t.Skip("no PowerShell flavor is installed")
	}
}

func TestSetupConfigPath(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	got, err := setupConfigPath("~/.bashrc")
	if err != nil {
		t.Fatalf("setupConfigPath() error = %v", err)
	}
	if got != filepath.Join(home, ".bashrc") {
		t.Fatalf("setupConfigPath() = %q", got)
	}
	if _, err := setupConfigPath("relative/path"); err == nil {
		t.Fatal("setupConfigPath() accepted a relative path")
	}
}

func TestAppendConfigurationLine(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config")
	if err := appendConfigurationLine(path, "line", nil); err != nil {
		t.Fatal(err)
	}
	if got := readFile(t, path); got != "line\n" {
		t.Fatalf("appendConfigurationLine(empty) = %q", got)
	}

	if err := os.WriteFile(path, []byte("first"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := appendConfigurationLine(path, "second", []byte("first")); err != nil {
		t.Fatal(err)
	}
	if got := readFile(t, path); got != "first\nsecond\n" {
		t.Fatalf("appendConfigurationLine(no newline) = %q", got)
	}

	if err := os.WriteFile(path, []byte("first\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := appendConfigurationLine(path, "third", []byte("first\n")); err != nil {
		t.Fatal(err)
	}
	if got := readFile(t, path); got != "first\nthird\n" {
		t.Fatalf("appendConfigurationLine(with newline) = %q", got)
	}

	if err := appendConfigurationLine(filepath.Join(dir, "missing", "config"), "line", nil); err == nil {
		t.Fatal("appendConfigurationLine() of a missing directory returned no error")
	}
}

func TestReplaceInstantModeLine(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config")
	legacy := "eval $(go-fuck --alias --enable-experimental-instant-mode)\nkeep\n"
	if err := os.WriteFile(path, []byte(legacy), 0o644); err != nil {
		t.Fatal(err)
	}
	content := "eval \"$(go-fuck --alias --enable-experimental-instant-mode)\""
	changed, err := replaceInstantModeLine(path, []byte(legacy), content)
	if err != nil {
		t.Fatalf("replaceInstantModeLine() error = %v", err)
	}
	if !changed {
		t.Fatal("replaceInstantModeLine() did not change the legacy line")
	}
	if got := readFile(t, path); !strings.Contains(got, content) || !strings.Contains(got, "keep") {
		t.Fatalf("replaceInstantModeLine() = %q", got)
	}

	changed, err = replaceInstantModeLine(path, []byte("nothing here\n"), content)
	if err != nil || changed {
		t.Fatalf("replaceInstantModeLine(nothing) = %t, %v", changed, err)
	}

	if _, err := replaceInstantModeLine(filepath.Join(t.TempDir(), "missing"), []byte(legacy), content); err == nil {
		t.Fatal("replaceInstantModeLine() of a missing file returned no error")
	}
}

func TestWellKnownPowerShellPath(t *testing.T) {
	t.Setenv("SystemRoot", "")
	if got := wellKnownPowerShellPath("powershell.exe"); got != "" {
		t.Fatalf("wellKnownPowerShellPath(no SystemRoot) = %q", got)
	}
	t.Setenv("SystemRoot", t.TempDir())
	if got := wellKnownPowerShellPath("powershell.exe"); got != "" {
		t.Fatalf("wellKnownPowerShellPath(missing binary) = %q", got)
	}
	if got := wellKnownPowerShellPath("pwsh"); got != "" {
		t.Fatalf("wellKnownPowerShellPath(pwsh) = %q", got)
	}
}

func TestSetupManualConfiguration(t *testing.T) {
	resetEntrypointsState(t)
	shells.Set(shells.NewGeneric())
	t.Cleanup(func() { shells.Set(shells.NewGeneric()) })
	if code := setup(&argumentparser.Args{}); code != 1 {
		t.Fatalf("setup() = %d, want 1", code)
	}
}
