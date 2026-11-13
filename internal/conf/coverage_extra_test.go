package conf

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

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
