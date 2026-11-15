package entrypoints

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/leafevans/go-fuck/v2/internal/conf"
	"github.com/leafevans/go-fuck/v2/internal/consts"
	"github.com/leafevans/go-fuck/v2/internal/logs"
)

func TestUnknownRuleSettings(t *testing.T) {
	if got := unknownRuleSettings(nil); got != nil {
		t.Fatalf("unknownRuleSettings(nil) = %q, want nil", got)
	}

	valid := conf.Default()
	valid.Rules = []string{consts.AllEnabled, "git_push"}
	valid.ExcludeRules = []string{"git_push"}
	valid.Priority = map[string]int{"git_push": 1}
	if got := unknownRuleSettings(valid); len(got) != 0 {
		t.Fatalf("unknownRuleSettings(valid) = %q, want none", got)
	}

	settings := conf.Default()
	settings.Rules = []string{consts.AllEnabled, "git_pushh", " "}
	settings.ExcludeRules = []string{"nope", ""}
	settings.Priority = map[string]int{"priority-typo": 10}
	want := []string{
		`exclude_rules "nope"`,
		`priority "priority-typo"`,
		`rules "git_pushh"`,
	}
	if got := unknownRuleSettings(settings); !reflect.DeepEqual(got, want) {
		t.Fatalf("unknownRuleSettings() = %q, want %q", got, want)
	}
}

func TestWarnUnknownRuleSettings(t *testing.T) {
	settings := conf.Default()
	settings.ExcludeRules = []string{"git_pushh"}
	buffer := &bytes.Buffer{}
	logs.SetOutput(buffer)
	t.Cleanup(func() {
		logs.SetOutput(os.Stderr)
	})

	warnUnknownRuleSettings(settings)
	output := buffer.String()
	if !strings.Contains(output, "Unknown rule names") || !strings.Contains(output, "git_pushh") {
		t.Fatalf("warnUnknownRuleSettings() output = %q, want the unknown rule name", output)
	}

	buffer.Reset()
	warnUnknownRuleSettings(conf.Default())
	if buffer.Len() != 0 {
		t.Fatalf("warnUnknownRuleSettings() warned for the default settings: %q", buffer.String())
	}
}

func TestConflictingRuleSettings(t *testing.T) {
	if got := conflictingRuleSettings(nil); got != nil {
		t.Fatalf("conflictingRuleSettings(nil) = %q, want nil", got)
	}
	if got := conflictingRuleSettings(conf.Default()); len(got) != 0 {
		t.Fatalf("conflictingRuleSettings(default) = %q, want none", got)
	}

	excludedSettings := conf.Default()
	excludedSettings.ExcludeRules = []string{"git_push", "unknown"}
	excludedSettings.Rules = []string{consts.AllEnabled, "git_push"}
	got := conflictingRuleSettings(excludedSettings)
	want := []string{`rules "git_push" also appears in exclude_rules`}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("conflictingRuleSettings() = %q, want %q", got, want)
	}

	prioritySettings := conf.Default()
	prioritySettings.ExcludeRules = []string{"git_push"}
	prioritySettings.Priority = map[string]int{"git_push": 10}
	got = conflictingRuleSettings(prioritySettings)
	want = []string{`priority "git_push" is ignored because the rule is excluded`}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("conflictingRuleSettings() = %q, want %q", got, want)
	}

	disabledSettings := conf.Default()
	disabledSettings.Rules = []string{"git_add"}
	disabledSettings.Priority = map[string]int{"git_push": 10}
	got = conflictingRuleSettings(disabledSettings)
	want = []string{`priority "git_push" is ignored because the rule is not enabled`}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("conflictingRuleSettings() = %q, want %q", got, want)
	}

	enabledSettings := conf.Default()
	enabledSettings.Rules = []string{consts.AllEnabled}
	enabledSettings.Priority = map[string]int{"git_push": 10}
	if got := conflictingRuleSettings(enabledSettings); len(got) != 0 {
		t.Fatalf("conflictingRuleSettings(enabled) = %q, want none", got)
	}
}

func TestWarnConflictingRuleSettings(t *testing.T) {
	settings := conf.Default()
	settings.ExcludeRules = []string{"git_push"}
	settings.Priority = map[string]int{"git_push": 10}
	buffer := &bytes.Buffer{}
	logs.SetOutput(buffer)
	t.Cleanup(func() {
		logs.SetOutput(os.Stderr)
	})

	warnConflictingRuleSettings(settings)
	if !strings.Contains(buffer.String(), "Conflicting rule settings") {
		t.Fatalf("warnConflictingRuleSettings() output = %q", buffer.String())
	}

	buffer.Reset()
	warnConflictingRuleSettings(conf.Default())
	if buffer.Len() != 0 {
		t.Fatalf("warnConflictingRuleSettings() warned for the default settings: %q", buffer.String())
	}
}

func TestDoctorWarnsAboutConflictingRuleSettings(t *testing.T) {
	resetEntrypointsState(t)
	settingsDir := conf.UserDir()
	writeTestFile(t, filepath.Join(settingsDir, "settings.json"),
		`{"exclude_rules": ["git_push"], "priority": {"git_push": 10}}`)

	code, stdout, _ := runMain(t, "go-fuck", "--doctor")
	if code != 0 {
		t.Fatalf("Main(--doctor) = %d, want 0", code)
	}
	for _, want := range []string{"inconsistent rule settings", "git_push"} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("doctor output %q does not contain %q", stdout, want)
		}
	}
}

func TestDoctorWarnsAboutUnknownRuleNames(t *testing.T) {
	resetEntrypointsState(t)
	settingsDir := conf.UserDir()
	writeTestFile(t, filepath.Join(settingsDir, "settings.json"), `{"exclude_rules": ["git_pushh"]}`)

	code, stdout, _ := runMain(t, "go-fuck", "--doctor")
	if code != 0 {
		t.Fatalf("Main(--doctor) = %d, want 0", code)
	}
	for _, want := range []string{"unknown rule names", "git_pushh"} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("doctor output %q does not contain %q", stdout, want)
		}
	}
}

func TestWarnInvalidSettings(t *testing.T) {
	settings := conf.Default()
	settings.InvalidValues = []string{"wait_command 0"}
	buffer := &bytes.Buffer{}
	logs.SetOutput(buffer)
	t.Cleanup(func() {
		logs.SetOutput(os.Stderr)
	})

	warnInvalidSettings(settings)
	if !strings.Contains(buffer.String(), "Invalid settings values") ||
		!strings.Contains(buffer.String(), "wait_command 0") {
		t.Fatalf("warnInvalidSettings() output = %q", buffer.String())
	}

	buffer.Reset()
	warnInvalidSettings(conf.Default())
	if buffer.Len() != 0 {
		t.Fatalf("warnInvalidSettings() warned for the default settings: %q", buffer.String())
	}
}

func TestDoctorWarnsAboutInvalidSettingsValues(t *testing.T) {
	resetEntrypointsState(t)
	settingsDir := conf.UserDir()
	writeTestFile(t, filepath.Join(settingsDir, "settings.json"), `{"wait_command": 0}`)

	code, stdout, _ := runMain(t, "go-fuck", "--doctor")
	if code != 0 {
		t.Fatalf("Main(--doctor) = %d, want 0", code)
	}
	for _, want := range []string{"invalid values", "wait_command 0"} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("doctor output %q does not contain %q", stdout, want)
		}
	}
}
