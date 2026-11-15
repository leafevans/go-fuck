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
