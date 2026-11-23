package conf

import (
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/leafevans/go-fuck/v3/internal/consts"
)

// isolateConfig points the user config directory at a temporary directory.
func isolateConfig(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("APPDATA", dir)
	t.Setenv("XDG_CONFIG_HOME", dir)
	t.Setenv("HOME", dir)
	Reset()
	t.Cleanup(Reset)
	return dir
}

// settingsFilePath returns the path of the settings file inside the isolated
// config directory.
func settingsFilePath(t *testing.T) string {
	t.Helper()
	base, err := os.UserConfigDir()
	if err != nil {
		t.Fatalf("os.UserConfigDir() error = %v", err)
	}
	return filepath.Join(base, "go-fuck", "settings.json")
}

func TestDefault(t *testing.T) {
	settings := Default()
	if len(settings.Rules) != 1 || settings.Rules[0] != consts.AllEnabled {
		t.Fatalf("Rules = %q, want [All rules enabled]", settings.Rules)
	}
	if !settings.RequireConfirmation || settings.NoColors || settings.Debug {
		t.Fatalf("unexpected flags: %+v", settings)
	}
	if settings.WaitCommand != 3 || settings.WaitSlowCommand != 15 || settings.NumCloseMatches != 3 {
		t.Fatalf("unexpected timeouts: %+v", settings)
	}
	if !settings.AlterHistory {
		t.Fatal("AlterHistory should be enabled by default")
	}
	if settings.Fzf {
		t.Fatal("Fzf should be disabled by default")
	}
	if settings.Env["GIT_TRACE"] != "1" {
		t.Fatalf("unexpected env: %q", settings.Env)
	}
}

func TestLoadCreatesSettingsFile(t *testing.T) {
	isolateConfig(t)
	if _, err := Load(&Args{}); err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	data, err := os.ReadFile(settingsFilePath(t))
	if err != nil {
		t.Fatalf("settings file was not created: %v", err)
	}
	if !strings.Contains(string(data), "require_confirmation") {
		t.Fatalf("unexpected settings file content: %q", data)
	}
}

func TestLoadReportsUnknownKeys(t *testing.T) {
	isolateConfig(t)
	if err := os.MkdirAll(filepath.Dir(settingsFilePath(t)), 0o755); err != nil {
		t.Fatal(err)
	}
	content := `{
		"_comment": "comment",
		"rules": ["git_push"],
		"rule": ["typo"],
		"wait_command": 9,
		"settings": {"fixlinecmd": "x"},
		"nope": true
	}`
	if err := os.WriteFile(settingsFilePath(t), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	settings, err := Load(nil)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if got := strings.Join(settings.UnknownKeys, ","); got != "nope,rule" {
		t.Fatalf("UnknownKeys = %q, want %q", got, "nope,rule")
	}
	if settings.WaitCommand != 9 || len(settings.Rules) != 1 || settings.Rules[0] != "git_push" {
		t.Fatalf("known settings were not applied: %+v", settings)
	}
}

func TestKnownSettingKeysMatchSchema(t *testing.T) {
	typ := reflect.TypeFor[fileSchema]()
	expected := map[string]bool{}
	for field := range typ.Fields() {
		name, _, _ := strings.Cut(field.Tag.Get("json"), ",")
		if name == "" || name == "-" {
			continue
		}
		expected[name] = true
	}
	for name := range expected {
		if !knownSettingKeys[name] {
			t.Fatalf("knownSettingKeys is missing the %q field", name)
		}
	}
	for name := range knownSettingKeys {
		if !expected[name] {
			t.Fatalf("knownSettingKeys contains %q, which is not a settings field", name)
		}
	}
}

func TestUnknownSettingKeysIgnoresEmptyName(t *testing.T) {
	got := unknownSettingKeys([]byte(`{"": true, "rule": true, "rules": []}`))
	if len(got) != 1 || got[0] != "rule" {
		t.Fatalf("unknownSettingKeys() = %q, want [rule]", got)
	}
}

func TestLoadFromFile(t *testing.T) {
	isolateConfig(t)
	if err := os.MkdirAll(filepath.Dir(settingsFilePath(t)), 0o755); err != nil {
		t.Fatal(err)
	}
	content := `{
		"rules": ["git_push", "quotation_marks"],
		"exclude_rules": ["cd_mkdir"],
		"wait_command": 7,
		"require_confirmation": false,
		"history_limit": 5,
		"slow_commands": ["lein"],
		"priority": {"git_push": 42},
		"settings": {"fixlinecmd": "{editor} {file} +{line}"},
		"fzf": true
	}`
	if err := os.WriteFile(settingsFilePath(t), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	settings, err := Load(nil)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(settings.Rules) != 2 || settings.Rules[0] != "git_push" {
		t.Fatalf("Rules = %q", settings.Rules)
	}
	if len(settings.ExcludeRules) != 1 || settings.ExcludeRules[0] != "cd_mkdir" {
		t.Fatalf("ExcludeRules = %q", settings.ExcludeRules)
	}
	if settings.WaitCommand != 7 || settings.RequireConfirmation || settings.HistoryLimit != 5 || !settings.Fzf {
		t.Fatalf("unexpected settings: %+v", settings)
	}
	if settings.Priority["git_push"] != 42 {
		t.Fatalf("Priority = %v", settings.Priority)
	}
	if got := settings.String("fixlinecmd", ""); got == "" {
		t.Fatal("fixlinecmd setting was not loaded")
	}
}

func TestLoadStrictSettingsRejectsUnknownKeys(t *testing.T) {
	isolateConfig(t)
	if err := os.MkdirAll(filepath.Dir(settingsFilePath(t)), 0o755); err != nil {
		t.Fatal(err)
	}
	content := `{"strict_settings": true, "wait_command": 9, "rule": ["typo"]}`
	if err := os.WriteFile(settingsFilePath(t), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := Load(nil)
	if err == nil || !strings.Contains(err.Error(), "unknown settings: rule") {
		t.Fatalf("Load() error = %v, want an unknown settings error", err)
	}
	if Current().WaitCommand != 3 {
		t.Fatalf("WaitCommand = %d, want the default because the file was rejected", Current().WaitCommand)
	}
	if got := strings.Join(Current().UnknownKeys, ","); got != "rule" {
		t.Fatalf("UnknownKeys = %q, want rule", got)
	}
}

func TestLoadStrictSettingsFromEnv(t *testing.T) {
	isolateConfig(t)
	if err := os.MkdirAll(filepath.Dir(settingsFilePath(t)), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(settingsFilePath(t), []byte(`{"rule": ["typo"]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GOFUCK_STRICT_SETTINGS", "true")

	if _, err := Load(nil); err == nil || !strings.Contains(err.Error(), "unknown settings: rule") {
		t.Fatalf("Load() error = %v, want an unknown settings error", err)
	}
}

func TestLoadStrictSettingsAcceptsKnownFile(t *testing.T) {
	isolateConfig(t)
	if err := os.MkdirAll(filepath.Dir(settingsFilePath(t)), 0o755); err != nil {
		t.Fatal(err)
	}
	content := `{"strict_settings": true, "wait_command": 9}`
	if err := os.WriteFile(settingsFilePath(t), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	settings, err := Load(nil)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if settings.WaitCommand != 9 || !settings.StrictSettings {
		t.Fatalf("unexpected settings: %+v", settings)
	}
	if len(settings.UnknownKeys) != 0 {
		t.Fatalf("UnknownKeys = %q, want none", settings.UnknownKeys)
	}
}

func TestLoadInvalidFile(t *testing.T) {
	isolateConfig(t)
	if err := os.MkdirAll(filepath.Dir(settingsFilePath(t)), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(settingsFilePath(t), []byte("{invalid"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GOFUCK_WAIT_COMMAND", "11")

	if _, err := Load(nil); err == nil {
		t.Fatal("Load() should report an invalid settings file")
	}
	if Current().WaitCommand != 11 {
		t.Fatalf("WaitCommand = %d, want the environment value 11", Current().WaitCommand)
	}
}

func TestEnvOverridesFile(t *testing.T) {
	isolateConfig(t)
	if err := os.MkdirAll(filepath.Dir(settingsFilePath(t)), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(settingsFilePath(t), []byte(`{"wait_command": 7}`), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GOFUCK_WAIT_COMMAND", "9")

	settings, err := Load(nil)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if settings.WaitCommand != 9 {
		t.Fatalf("WaitCommand = %d, want 9", settings.WaitCommand)
	}
}

// writeSettingsFile writes content to the isolated settings file.
func writeSettingsFile(t *testing.T, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(settingsFilePath(t)), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(settingsFilePath(t), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestLoadIgnoresUnknownFields(t *testing.T) {
	isolateConfig(t)
	writeSettingsFile(t, `{
		"wait_command": 9,
		"future_option": {"nested": true},
		"settings": {"fixlinecmd": "{editor} {file}"}
	}`)

	settings, err := Load(nil)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if settings.WaitCommand != 9 {
		t.Fatalf("WaitCommand = %d, want 9", settings.WaitCommand)
	}
	if !settings.RequireConfirmation {
		t.Fatal("RequireConfirmation = false, want the default true")
	}
	if got := settings.String("fixlinecmd", ""); got != "{editor} {file}" {
		t.Fatalf("fixlinecmd = %q, want the settings map value", got)
	}
}

func TestLoadRejectsWrongFieldTypes(t *testing.T) {
	isolateConfig(t)
	writeSettingsFile(t, `{"wait_command": "many"}`)

	if _, err := Load(nil); err == nil {
		t.Fatal("Load() should report a settings value of the wrong type")
	}
}

func TestLoadMissingFieldsKeepDefaults(t *testing.T) {
	isolateConfig(t)
	writeSettingsFile(t, "{}")

	settings, err := Load(nil)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if settings.WaitCommand != 3 || !settings.RequireConfirmation || settings.NoColors {
		t.Fatalf("unexpected defaults: %+v", settings)
	}
}

func TestLoadEnv(t *testing.T) {
	isolateConfig(t)
	t.Setenv("GOFUCK_RULES", "DEFAULT_RULES:git_push")
	t.Setenv("GOFUCK_EXCLUDE_RULES", "cd_mkdir")
	t.Setenv("GOFUCK_WAIT_COMMAND", "4")
	t.Setenv("GOFUCK_REQUIRE_CONFIRMATION", "FALSE")
	t.Setenv("GOFUCK_NO_COLORS", "true")
	t.Setenv("GOFUCK_DEBUG", "true")
	t.Setenv("GOFUCK_HISTORY_LIMIT", "10")
	t.Setenv("GOFUCK_ALTER_HISTORY", "false")
	t.Setenv("GOFUCK_WAIT_SLOW_COMMAND", "20")
	t.Setenv("GOFUCK_SLOW_COMMANDS", "lein:gradle")
	t.Setenv("GOFUCK_REPEAT", "true")
	t.Setenv("GOFUCK_INSTANT_MODE", "true")
	t.Setenv("GOFUCK_FZF", "true")
	t.Setenv("GOFUCK_NUM_CLOSE_MATCHES", "5")
	t.Setenv("GOFUCK_EXCLUDED_SEARCH_PATH_PREFIXES", "/mnt")
	t.Setenv("GOFUCK_PRIORITY", "git_push=1:other=2:invalid")

	settings, err := Load(nil)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(settings.Rules) != 2 || settings.Rules[0] != consts.AllEnabled || settings.Rules[1] != "git_push" {
		t.Fatalf("Rules = %q", settings.Rules)
	}
	if len(settings.ExcludeRules) != 1 || settings.ExcludeRules[0] != "cd_mkdir" {
		t.Fatalf("ExcludeRules = %q", settings.ExcludeRules)
	}
	if settings.WaitCommand != 4 {
		t.Fatalf("WaitCommand = %d", settings.WaitCommand)
	}
	if settings.RequireConfirmation {
		t.Fatal("RequireConfirmation = true, want false")
	}
	if !settings.NoColors || !settings.Debug || !settings.Repeat || !settings.InstantMode || !settings.Fzf {
		t.Fatalf("unexpected flags: %+v", settings)
	}
	if settings.HistoryLimit != 10 || settings.WaitSlowCommand != 20 || settings.NumCloseMatches != 5 {
		t.Fatalf("unexpected numeric settings: %+v", settings)
	}
	if settings.AlterHistory {
		t.Fatal("AlterHistory = true, want false")
	}
	if len(settings.SlowCommands) != 2 || settings.SlowCommands[1] != "gradle" {
		t.Fatalf("SlowCommands = %q", settings.SlowCommands)
	}
	if len(settings.ExcludedSearchPathPrefixes) != 1 {
		t.Fatalf("ExcludedSearchPathPrefixes = %q", settings.ExcludedSearchPathPrefixes)
	}
	if settings.Priority["git_push"] != 1 || settings.Priority["other"] != 2 {
		t.Fatalf("Priority = %v", settings.Priority)
	}
}

func TestLoadLegacyEnv(t *testing.T) {
	isolateConfig(t)
	t.Setenv("THEFUCK_WAIT_COMMAND", "6")
	t.Setenv("THEFUCK_HISTORY_LIMIT", "3")

	settings, err := Load(nil)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if settings.WaitCommand != 6 || settings.HistoryLimit != 3 {
		t.Fatalf("legacy environment variables were ignored: %+v", settings)
	}
}

func TestLoadRulesWithoutDefault(t *testing.T) {
	isolateConfig(t)
	t.Setenv("GOFUCK_RULES", "git_push:cd_mkdir")

	settings, err := Load(nil)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(settings.Rules) != 2 || settings.Rules[0] != "git_push" {
		t.Fatalf("Rules = %q", settings.Rules)
	}
}

func TestLoadArgs(t *testing.T) {
	isolateConfig(t)
	settings, err := Load(&Args{Yes: true, Debug: true, Repeat: true})
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if settings.RequireConfirmation || !settings.Debug || !settings.Repeat {
		t.Fatalf("arguments were not applied: %+v", settings)
	}
}

func TestAlias(t *testing.T) {
	t.Setenv("TF_ALIAS", "please")
	if got := Alias(); got != "please" {
		t.Fatalf("Alias() = %q, want please", got)
	}
	t.Setenv("TF_ALIAS", "")
	if got := Alias(); got != "fuck" {
		t.Fatalf("Alias() = %q, want fuck", got)
	}
}

func TestStringExtra(t *testing.T) {
	settings := Default()
	if got := settings.String("missing", "default"); got != "default" {
		t.Fatalf("String() = %q, want default", got)
	}
	settings.SetExtra("key", "value")
	if got := settings.String("key", "default"); got != "value" {
		t.Fatalf("String() = %q, want value", got)
	}
}

func TestShellFromEnv(t *testing.T) {
	t.Setenv("TF_SHELL", "zsh")
	if got := ShellFromEnv(); got != "zsh" {
		t.Fatalf("ShellFromEnv() = %q, want zsh", got)
	}
}

func TestLoadExcludedSearchPathPrefixesSeparator(t *testing.T) {
	isolateConfig(t)
	prefixes := []string{"/snap", "/mnt"}
	if runtime.GOOS == "windows" {
		// A colon is part of these paths, so they must not be split on it.
		prefixes = []string{`C:\Tools`, `C:\Other`}
	}
	t.Setenv("GOFUCK_EXCLUDED_SEARCH_PATH_PREFIXES",
		strings.Join(prefixes, string(os.PathListSeparator)))

	settings, err := Load(nil)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(settings.ExcludedSearchPathPrefixes) != len(prefixes) {
		t.Fatalf("ExcludedSearchPathPrefixes = %q, want %q",
			settings.ExcludedSearchPathPrefixes, prefixes)
	}
	for index, prefix := range prefixes {
		if settings.ExcludedSearchPathPrefixes[index] != prefix {
			t.Fatalf("ExcludedSearchPathPrefixes[%d] = %q, want %q", index,
				settings.ExcludedSearchPathPrefixes[index], prefix)
		}
	}
}

func TestUserDirFallback(t *testing.T) {
	t.Setenv("APPDATA", "")
	t.Setenv("XDG_CONFIG_HOME", "")
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	dir := UserDir()
	if !strings.Contains(dir, "go-fuck") {
		t.Fatalf("UserDir() = %q, want a go-fuck directory", dir)
	}
}

func TestExtraSettings(t *testing.T) {
	settings := Default()
	if got := settings.String("missing", "fallback"); got != "fallback" {
		t.Fatalf("String(missing) = %q", got)
	}
	settings.SetExtra("fixlinecmd", "value")
	if got := settings.String("fixlinecmd", "fallback"); got != "value" {
		t.Fatalf("String(fixlinecmd) = %q", got)
	}

	empty := &Settings{}
	empty.SetExtra("key", "value")
	if got := empty.String("key", ""); got != "value" {
		t.Fatalf("String(key) = %q", got)
	}
}

func TestStringWithoutExtraSettings(t *testing.T) {
	settings := &Settings{}
	if got := settings.String("missing", "fallback"); got != "fallback" {
		t.Fatalf("String() = %q, want fallback", got)
	}
}

func TestUserDirWithoutConfigDirOrHome(t *testing.T) {
	t.Setenv("APPDATA", "")
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("HOME", "")
	t.Setenv("USERPROFILE", "")
	t.Setenv("HOMEDRIVE", "")
	t.Setenv("HOMEPATH", "")
	if got := UserDir(); got != ".go-fuck" {
		t.Fatalf("UserDir() = %q, want .go-fuck", got)
	}
}

func TestUnknownSettingKeysInvalidJSON(t *testing.T) {
	if got := unknownSettingKeys([]byte("{invalid")); got != nil {
		t.Fatalf("unknownSettingKeys() = %q, want nil", got)
	}
}

func TestLoadFileWithoutUserDir(t *testing.T) {
	settings := Default()
	settings.UserDir = ""
	if err := settings.loadFile(); err != nil {
		t.Fatalf("loadFile() error = %v", err)
	}
}

func TestLoadFileCreateDirFailure(t *testing.T) {
	blocker := filepath.Join(t.TempDir(), "blocker")
	if err := os.WriteFile(blocker, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	settings := Default()
	settings.UserDir = filepath.Join(blocker, "go-fuck")
	err := settings.loadFile()
	if err == nil || !strings.Contains(err.Error(), "create config dir") {
		t.Fatalf("loadFile() error = %v, want a create config dir error", err)
	}
}

func TestLoadFileReadFailure(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "settings.json"), 0o755); err != nil {
		t.Fatal(err)
	}
	settings := Default()
	settings.UserDir = dir
	err := settings.loadFile()
	if err == nil || !strings.Contains(err.Error(), "read settings file") {
		t.Fatalf("loadFile() error = %v, want a read settings file error", err)
	}
}

func TestLoadFileWriteDefaultFailure(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("directory write permissions are not enforced on Windows")
	}
	userDir := filepath.Join(t.TempDir(), "go-fuck")
	if err := os.MkdirAll(userDir, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(userDir, 0o755) })
	if err := os.Chmod(userDir, 0o500); err != nil {
		t.Fatal(err)
	}

	settings := Default()
	settings.UserDir = userDir
	err := settings.loadFile()
	if err == nil {
		t.Skip("the test process can write to a directory without write permission")
	}
	if !strings.Contains(err.Error(), "write default settings file") {
		t.Fatalf("loadFile() error = %v, want a write default settings error", err)
	}
}

func TestApplyFileAllFields(t *testing.T) {
	rules := []string{"git_push"}
	excludeRules := []string{"cd_mkdir"}
	waitCommand := 7
	requireConfirmation := false
	noColors := true
	debug := true
	priority := map[string]int{"git_push": 5}
	historyLimit := 4
	alterHistory := false
	waitSlowCommand := 20
	slowCommands := []string{"lein"}
	repeat := true
	instantMode := true
	fzf := true
	strictSettings := true
	numCloseMatches := 9
	env := map[string]string{"LC_ALL": "C"}
	excludedSearchPathPrefixes := []string{"/mnt"}

	settings := Default()
	settings.applyFile(&fileSchema{
		Rules:                      &rules,
		ExcludeRules:               &excludeRules,
		WaitCommand:                &waitCommand,
		RequireConfirmation:        &requireConfirmation,
		NoColors:                   &noColors,
		Debug:                      &debug,
		Priority:                   &priority,
		HistoryLimit:               &historyLimit,
		AlterHistory:               &alterHistory,
		WaitSlowCommand:            &waitSlowCommand,
		SlowCommands:               &slowCommands,
		Repeat:                     &repeat,
		InstantMode:                &instantMode,
		Fzf:                        &fzf,
		StrictSettings:             &strictSettings,
		NumCloseMatches:            &numCloseMatches,
		Env:                        &env,
		ExcludedSearchPathPrefixes: &excludedSearchPathPrefixes,
	})

	if len(settings.Rules) != 1 || settings.Rules[0] != "git_push" {
		t.Fatalf("Rules = %q", settings.Rules)
	}
	if len(settings.ExcludeRules) != 1 || settings.ExcludeRules[0] != "cd_mkdir" {
		t.Fatalf("ExcludeRules = %q", settings.ExcludeRules)
	}
	if settings.WaitCommand != 7 || settings.HistoryLimit != 4 || settings.WaitSlowCommand != 20 || settings.NumCloseMatches != 9 {
		t.Fatalf("unexpected numeric settings: %+v", settings)
	}
	if settings.RequireConfirmation || !settings.NoColors || !settings.Debug || settings.AlterHistory {
		t.Fatalf("unexpected flags: %+v", settings)
	}
	if !settings.Repeat || !settings.InstantMode || !settings.Fzf || !settings.StrictSettings {
		t.Fatalf("unexpected flags: %+v", settings)
	}
	if settings.Priority["git_push"] != 5 || settings.Env["LC_ALL"] != "C" {
		t.Fatalf("unexpected maps: %+v", settings)
	}
	if len(settings.SlowCommands) != 1 || len(settings.ExcludedSearchPathPrefixes) != 1 {
		t.Fatalf("unexpected slices: %+v", settings)
	}
}

func TestLoadPriorityEnvInvalidNumber(t *testing.T) {
	isolateConfig(t)
	t.Setenv("GOFUCK_PRIORITY", "git_push=not-a-number")

	settings, err := Load(nil)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(settings.Priority) != 0 {
		t.Fatalf("Priority = %v, want no entries for an invalid number", settings.Priority)
	}
}
