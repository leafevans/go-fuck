package entrypoints

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/leafevans/go-fuck/v3/internal/argumentparser"
	"github.com/leafevans/go-fuck/v3/internal/shells"

	"github.com/leafevans/go-fuck/v3/internal/conf"
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

	code, stdout, stderr := runMain(t, "go-fuck", "--setup")
	if code != 0 {
		t.Fatalf("Main(--setup) = %d, want 0; stderr = %s", code, stderr)
	}
	if !strings.Contains(stdout, "now configured") {
		t.Fatalf("stdout = %q, want a success message", stdout)
	}
	if stderr != "" {
		t.Fatalf("stderr = %q, want empty", stderr)
	}
	content := readFile(t, configPath)
	if !strings.Contains(content, "export EXISTING=1") {
		t.Fatalf("existing configuration was lost: %q", content)
	}
	if got := strings.Count(content, "go-fuck --alias"); got != 1 {
		t.Fatalf("configuration contains %d alias lines, want 1: %q", got, content)
	}

	code, stdout, stderr = runMain(t, "go-fuck", "--setup")
	if code != 0 {
		t.Fatalf("second Main(--setup) = %d, want 0", code)
	}
	if !strings.Contains(stdout, "already configured") {
		t.Fatalf("stdout = %q, want an already configured message", stdout)
	}
	if stderr != "" {
		t.Fatalf("stderr = %q, want empty", stderr)
	}
	if got := strings.Count(readFile(t, configPath), "go-fuck --alias"); got != 1 {
		t.Fatalf("configuration contains %d alias lines after the second run, want 1", got)
	}
}

func TestMainSetupMigratesRegularAliasToAutoCapture(t *testing.T) {
	configDir := resetEntrypointsState(t)
	t.Setenv("TF_SHELL", "bash")
	configPath := filepath.Join(configDir, ".bashrc")
	regular := `eval "$(go-fuck --alias)"`
	if err := os.WriteFile(configPath, []byte(regular+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	code, stdout, stderr := runMain(t, "go-fuck", "--setup")
	if code != 0 || stderr != "" || !strings.Contains(stdout, "now configured") {
		t.Fatalf("setup = %d, %q, %q", code, stdout, stderr)
	}
	content := readFile(t, configPath)
	if strings.Contains(content, regular+"\n") ||
		!strings.Contains(content, `eval "$(go-fuck --alias --capture-mode=auto --enable-experimental-instant-mode)"`) {
		t.Fatalf("alias was not migrated to auto capture: %q", content)
	}
	if strings.Count(content, "go-fuck --alias") != 1 {
		t.Fatalf("migration duplicated the alias: %q", content)
	}
}

func TestMainSetupCaptureModeRerunKeepsRegularAlias(t *testing.T) {
	configDir := resetEntrypointsState(t)
	t.Setenv("TF_SHELL", "bash")
	configPath := filepath.Join(configDir, ".bashrc")
	if err := os.WriteFile(configPath, []byte("# existing\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	code, _, stderr := runMain(t, "go-fuck", "--setup", "--capture-mode", "rerun")
	if code != 0 || stderr != "" {
		t.Fatalf("setup = %d, %q", code, stderr)
	}
	content := readFile(t, configPath)
	if !strings.Contains(content, `eval "$(go-fuck --alias --capture-mode=rerun)"`) ||
		strings.Contains(content, "--enable-experimental-instant-mode") {
		t.Fatalf("rerun setup wrote the wrong alias: %q", content)
	}
}

func TestMainSetupRejectsAmbiguousInitializationLines(t *testing.T) {
	configDir := resetEntrypointsState(t)
	t.Setenv("TF_SHELL", "bash")
	configPath := filepath.Join(configDir, ".bashrc")
	regular := `eval "$(go-fuck --alias)"`
	original := regular + "\n" + regular + "\n"
	if err := os.WriteFile(configPath, []byte(original), 0o644); err != nil {
		t.Fatal(err)
	}

	code, _, stderr := runMain(t, "go-fuck", "--setup")
	if code != 1 || !strings.Contains(stderr, "multiple go-fuck initialization lines") {
		t.Fatalf("setup = %d, stderr=%q", code, stderr)
	}
	if content := readFile(t, configPath); content != original {
		t.Fatalf("ambiguous configuration was modified: %q", content)
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
	if !strings.Contains(content, "eval \"$(go-fuck --alias --capture-mode=instant --enable-experimental-instant-mode)\"") {
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

	code, stdout, stderr := runMain(t, "go-fuck", "--setup")
	if code != 0 {
		t.Fatalf("Main(--setup) = %d, want 0", code)
	}
	if !strings.Contains(stdout, "Add the following line") || !strings.Contains(stdout, "go-fuck --alias") {
		t.Fatalf("stdout = %q, want manual configuration instructions", stdout)
	}
	if stderr != "" {
		t.Fatalf("stderr = %q, want empty", stderr)
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

	code, stdout, stderr := runMain(t, "go-fuck", "--setup")
	if code != 0 {
		t.Fatalf("Main(--setup) = %d, want 0; stderr = %s", code, stderr)
	}
	if got := strings.Count(stdout, "now configured"); got != 2 {
		t.Fatalf("stdout = %q, want two success messages", stdout)
	}
	if !strings.Contains(stdout, existingProfile) || !strings.Contains(stdout, missingProfile) {
		t.Fatalf("stdout = %q, want both profile paths", stdout)
	}
	if stderr != "" {
		t.Fatalf("stderr = %q, want empty", stderr)
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

	code, stdout, stderr = runMain(t, "go-fuck", "--setup")
	if code != 0 {
		t.Fatalf("second Main(--setup) = %d, want 0", code)
	}
	if !strings.Contains(stdout, "already configured") {
		t.Fatalf("stdout = %q, want an already configured message", stdout)
	}
	if stderr != "" {
		t.Fatalf("stderr = %q, want empty", stderr)
	}
	if got := strings.Count(readFile(t, existingProfile), "go-fuck --alias"); got != 1 {
		t.Fatalf("existing profile contains %d alias lines after the second run, want 1", got)
	}
}

func TestMainSetupPowerShellUsesExistingAllHostsProfile(t *testing.T) {
	configDir := resetEntrypointsState(t)
	t.Setenv("TF_SHELL", "powershell")
	dir := filepath.Join(configDir, "Documents", "PowerShell")
	allHosts := filepath.Join(dir, "profile.ps1")
	currentHost := filepath.Join(dir, "Microsoft.PowerShell_profile.ps1")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(allHosts, []byte("iex \"$(go-fuck --alias)\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	stubPowerShellProfiles(t, []string{currentHost})

	code, stdout, stderr := runMain(t, "go-fuck", "--setup")
	if code != 0 {
		t.Fatalf("Main(--setup) = %d, want 0; stderr = %s", code, stderr)
	}
	if !strings.Contains(stdout, allHosts) || strings.Contains(stdout, "now configured for "+currentHost) {
		t.Fatalf("stdout = %q, want the existing all-hosts profile", stdout)
	}
	if _, err := os.Stat(currentHost); !os.IsNotExist(err) {
		t.Fatalf("setup created the current-host profile: %v", err)
	}
	if got := strings.Count(readFile(t, allHosts), "go-fuck --alias"); got != 1 {
		t.Fatalf("all-hosts profile contains %d alias lines, want 1", got)
	}
}

func TestMainSetupPowerShellRejectsMultipleProfileInitializations(t *testing.T) {
	configDir := resetEntrypointsState(t)
	t.Setenv("TF_SHELL", "powershell")
	dir := filepath.Join(configDir, "Documents", "PowerShell")
	allHosts := filepath.Join(dir, "profile.ps1")
	currentHost := filepath.Join(dir, "Microsoft.PowerShell_profile.ps1")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	content := "iex \"$(go-fuck --alias)\"\n"
	for _, path := range []string{allHosts, currentHost} {
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	stubPowerShellProfiles(t, []string{currentHost})

	code, _, stderr := runMain(t, "go-fuck", "--setup")
	if code != 1 || !strings.Contains(stderr, "multiple go-fuck initializations") {
		t.Fatalf("Main(--setup) = %d, stderr = %q, want a duplicate warning", code, stderr)
	}
	if got := readFile(t, allHosts); got != content {
		t.Fatalf("all-hosts profile was modified: %q", got)
	}
	if got := readFile(t, currentHost); got != content {
		t.Fatalf("current-host profile was modified: %q", got)
	}
}

func TestMainSetupPowerShellRejectsDuplicateLinesInOneProfile(t *testing.T) {
	configDir := resetEntrypointsState(t)
	t.Setenv("TF_SHELL", "powershell")
	currentHost := filepath.Join(configDir, "Documents", "PowerShell", "Microsoft.PowerShell_profile.ps1")
	if err := os.MkdirAll(filepath.Dir(currentHost), 0o755); err != nil {
		t.Fatal(err)
	}
	content := "iex \"$(go-fuck --alias)\"\niex \"$(go-fuck --alias)\"\n"
	if err := os.WriteFile(currentHost, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	stubPowerShellProfiles(t, []string{currentHost})

	code, _, stderr := runMain(t, "go-fuck", "--setup")
	if code != 1 || !strings.Contains(stderr, "multiple go-fuck initializations") {
		t.Fatalf("Main(--setup) = %d, stderr = %q, want a duplicate warning", code, stderr)
	}
	if got := readFile(t, currentHost); got != content {
		t.Fatalf("current-host profile was modified: %q", got)
	}
}

func TestMainSetupPowerShellRejectsCustomInitialization(t *testing.T) {
	configDir := resetEntrypointsState(t)
	t.Setenv("TF_SHELL", "powershell")
	dir := filepath.Join(configDir, "Documents", "PowerShell")
	allHosts := filepath.Join(dir, "profile.ps1")
	currentHost := filepath.Join(dir, "Microsoft.PowerShell_profile.ps1")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	custom := "$alias = Get-CachedInit -Name 'gofuck' -Generate { & go-fuck --alias }\n"
	if err := os.WriteFile(allHosts, []byte(custom), 0o644); err != nil {
		t.Fatal(err)
	}
	stubPowerShellProfiles(t, []string{currentHost})

	code, _, stderr := runMain(t, "go-fuck", "--setup")
	if code != 1 || !strings.Contains(stderr, "custom go-fuck initialization") {
		t.Fatalf("Main(--setup) = %d, stderr = %q, want a custom initialization warning", code, stderr)
	}
	if got := readFile(t, allHosts); got != custom {
		t.Fatalf("custom profile was modified: %q", got)
	}
	if _, err := os.Stat(currentHost); !os.IsNotExist(err) {
		t.Fatalf("setup created the current-host profile: %v", err)
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
	if !strings.Contains(content, "--capture-mode=instant") ||
		!strings.Contains(content, "--enable-experimental-instant-mode") {
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

	code, stdout, stderr := runMain(t, "go-fuck", "--setup", "--enable-experimental-instant-mode")
	if code != 0 {
		t.Fatalf("Main(--setup --enable-experimental-instant-mode) = %d, want 0; stderr = %s", code, stderr)
	}
	if !strings.Contains(stderr, "instant capture is not supported") {
		t.Fatalf("stderr = %q, want an instant capture warning", stderr)
	}
	if !strings.Contains(stdout, "now configured") {
		t.Fatalf("stdout = %q, want the regular alias success message", stdout)
	}
	content := readFile(t, configPath)
	if strings.Contains(content, "--enable-experimental-instant-mode") {
		t.Fatalf("unsupported shell was configured with instant mode: %q", content)
	}
	if !strings.Contains(content, "go-fuck --alias --capture-mode=rerun | source") {
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
	if content := readFile(t, profile); !strings.Contains(content, "--capture-mode=instant") ||
		!strings.Contains(content, "--enable-experimental-instant-mode") {
		t.Fatalf("profile does not enable instant mode: %q", content)
	}
}

func TestMainSetupMissingConfigFile(t *testing.T) {
	resetEntrypointsState(t)
	t.Setenv("TF_SHELL", "zsh")

	code, stdout, stderr := runMain(t, "go-fuck", "--setup")
	if code != 0 {
		t.Fatalf("Main(--setup) = %d, want 0", code)
	}
	if !strings.Contains(stdout, "~/.zshrc") {
		t.Fatalf("stdout = %q, want the zsh configuration path", stdout)
	}
	if stderr != "" {
		t.Fatalf("stderr = %q, want empty", stderr)
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

// setupTestShell is a fake shell whose configuration can be controlled by the
// test.
type setupTestShell struct {
	*shells.Generic
	name       string
	configured *shells.Configuration
}

func (s setupTestShell) Name() string         { return s.name }
func (s setupTestShell) FriendlyName() string { return s.name }
func (s setupTestShell) HowToConfigure() *shells.Configuration {
	return s.configured
}

func TestSetupManualConfigurationPaths(t *testing.T) {
	resetEntrypointsState(t)
	t.Cleanup(func() { shells.Set(shells.NewGeneric()) })

	cases := []struct {
		name string
		path string
		want int
	}{
		{"shell variable path", "$PROFILE", 0},
		{"relative path", "relative/config", 0},
		{"missing absolute path", filepath.Join(t.TempDir(), "missing", "config"), 0},
		{"directory path", t.TempDir(), 1},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			shells.Set(setupTestShell{
				Generic: shells.NewGeneric(),
				name:    "test",
				configured: &shells.Configuration{
					Content:                   "eval \"$(go-fuck --alias)\"",
					Path:                      testCase.path,
					Reload:                    "reload",
					CanConfigureAutomatically: true,
				},
			})
			if code := setup(&argumentparser.Args{}); code != testCase.want {
				t.Fatalf("setup() = %d, want %d for %s", code, testCase.want, testCase.path)
			}
		})
	}
}

func TestSetupReportsSettingsLoadFailure(t *testing.T) {
	resetEntrypointsState(t)
	settingsPath := filepath.Join(conf.UserDir(), "settings.json")
	if err := os.MkdirAll(filepath.Dir(settingsPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(settingsPath, []byte("{invalid"), 0o644); err != nil {
		t.Fatal(err)
	}

	code, _, stderr := runMain(t, "go-fuck", "--setup")
	if code != 0 {
		t.Fatalf("Main(--setup) = %d, want 0; stderr = %s", code, stderr)
	}
	if !strings.Contains(stderr, "can't load settings") {
		t.Fatalf("stderr = %q, want a settings load failure", stderr)
	}
}

func TestSetupAppendFailure(t *testing.T) {
	configDir := resetEntrypointsState(t)
	configPath := filepath.Join(configDir, ".bashrc")
	if err := os.WriteFile(configPath, []byte("export EXISTING=1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	makeReadOnly(t, configPath)

	code, _, stderr := runMain(t, "go-fuck", "--setup")
	if code == 0 {
		t.Skip("the read only configuration file could be written (running as root?)")
	}
	if code != 1 || !strings.Contains(stderr, "can't update") {
		t.Fatalf("Main(--setup) = %d, stderr = %q, want a failed update", code, stderr)
	}
}

func TestSetupReplaceInstantModeLineFailure(t *testing.T) {
	configDir := resetEntrypointsState(t)
	configPath := filepath.Join(configDir, ".bashrc")
	legacy := "eval $(go-fuck --alias --enable-experimental-instant-mode)\n"
	if err := os.WriteFile(configPath, []byte(legacy), 0o644); err != nil {
		t.Fatal(err)
	}
	makeReadOnly(t, configPath)

	code, _, stderr := runMain(t, "go-fuck", "--setup", "--enable-experimental-instant-mode")
	if code == 0 {
		t.Skip("the read only configuration file could be written (running as root?)")
	}
	if code != 1 || !strings.Contains(stderr, "can't update") {
		t.Fatalf("Main(--setup) = %d, stderr = %q, want a failed update", code, stderr)
	}
}

func TestSetupPowerShellFailures(t *testing.T) {
	t.Run("profile is a directory", func(t *testing.T) {
		resetEntrypointsState(t)
		t.Setenv("TF_SHELL", "powershell")
		stubPowerShellProfiles(t, []string{t.TempDir()})

		code, _, stderr := runMain(t, "go-fuck", "--setup")
		if code != 1 || !strings.Contains(stderr, "can't read") {
			t.Fatalf("Main(--setup) = %d, stderr = %q, want a failed read", code, stderr)
		}
	})

	t.Run("profile directory cannot be created", func(t *testing.T) {
		resetEntrypointsState(t)
		t.Setenv("TF_SHELL", "powershell")
		var profile string
		if runtime.GOOS == "windows" {
			// A path below a regular file is reported as a missing file on
			// Windows, so the profile lookup passes and creating its directory
			// fails.
			blocker := filepath.Join(t.TempDir(), "blocker")
			if err := os.WriteFile(blocker, nil, 0o644); err != nil {
				t.Fatal(err)
			}
			profile = filepath.Join(blocker, "Documents", "profile.ps1")
		} else {
			// A read only parent directory makes the profile lookup report a
			// missing file and creating its directory fail.
			parent := t.TempDir()
			if err := os.Chmod(parent, 0o500); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = os.Chmod(parent, 0o755) })
			profile = filepath.Join(parent, "Documents", "profile.ps1")
		}
		stubPowerShellProfiles(t, []string{profile})

		code, _, stderr := runMain(t, "go-fuck", "--setup")
		if code == 0 {
			t.Skip("the missing profile directory could be created (running as root?)")
		}
		if code != 1 || !strings.Contains(stderr, "can't create") {
			t.Fatalf("Main(--setup) = %d, stderr = %q, want a failed directory creation", code, stderr)
		}
	})

	t.Run("profile cannot be appended", func(t *testing.T) {
		resetEntrypointsState(t)
		t.Setenv("TF_SHELL", "powershell")
		profile := filepath.Join(t.TempDir(), "profile.ps1")
		if err := os.WriteFile(profile, nil, 0o644); err != nil {
			t.Fatal(err)
		}
		makeReadOnly(t, profile)
		stubPowerShellProfiles(t, []string{profile})

		code, _, stderr := runMain(t, "go-fuck", "--setup")
		if code == 0 {
			t.Skip("the read only profile could be written (running as root?)")
		}
		if code != 1 || !strings.Contains(stderr, "can't update") {
			t.Fatalf("Main(--setup) = %d, stderr = %q, want a failed update", code, stderr)
		}
	})
}

func TestPowerShellProfileFakeExecutable(t *testing.T) {
	cases := []struct {
		name     string
		mode     string
		wantPath string
		wantOK   bool
	}{
		{name: "reports a profile", mode: "profile", wantPath: "/tmp/go-fuck-fake-profile.ps1", wantOK: true},
		{name: "fails", mode: "fail"},
		{name: "reports an empty profile", mode: "profile-empty"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			resetEntrypointsState(t)
			installFakeExecutable(t, "pwsh", testCase.mode)

			path, ok := powerShellProfile("pwsh")
			if ok != testCase.wantOK {
				t.Fatalf("powerShellProfile() = %q, %t, want ok = %t", path, ok, testCase.wantOK)
			}
			if path != testCase.wantPath {
				t.Fatalf("powerShellProfile() = %q, want %q", path, testCase.wantPath)
			}
		})
	}
}
