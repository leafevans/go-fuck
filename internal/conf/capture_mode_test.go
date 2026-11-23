package conf

import (
	"strings"
	"testing"
)

func TestCaptureModeDefaultsToAuto(t *testing.T) {
	isolateConfig(t)
	settings, err := Load(nil)
	if err != nil {
		t.Fatal(err)
	}
	if settings.CaptureMode != "auto" {
		t.Fatalf("CaptureMode = %q, want auto", settings.CaptureMode)
	}
}

func TestCaptureModePrecedenceAndLegacyMapping(t *testing.T) {
	isolateConfig(t)
	writeSettingsFile(t, `{"capture_mode": "instant", "instant_mode": false}`)
	settings, err := Load(nil)
	if err != nil {
		t.Fatal(err)
	}
	if settings.CaptureMode != "instant" {
		t.Fatalf("file CaptureMode = %q, want instant", settings.CaptureMode)
	}

	t.Setenv("GOFUCK_INSTANT_MODE", "false")
	settings, err = Load(nil)
	if err != nil {
		t.Fatal(err)
	}
	if settings.CaptureMode != "instant" {
		t.Fatalf("capture_mode should beat legacy env, got %q", settings.CaptureMode)
	}

	t.Setenv("GOFUCK_CAPTURE_MODE", "rerun")
	settings, err = Load(nil)
	if err != nil {
		t.Fatal(err)
	}
	if settings.CaptureMode != "rerun" {
		t.Fatalf("environment CaptureMode = %q, want rerun", settings.CaptureMode)
	}

	settings, err = Load(&Args{CaptureMode: "instant"})
	if err != nil {
		t.Fatal(err)
	}
	if settings.CaptureMode != "instant" {
		t.Fatalf("CLI CaptureMode = %q, want instant", settings.CaptureMode)
	}
}

func TestLegacyInstantModeMapsToCaptureMode(t *testing.T) {
	for _, tc := range []struct {
		name string
		file string
		env  string
		want string
	}{
		{name: "file true", file: `{"instant_mode": true}`, want: "auto"},
		{name: "file false", file: `{"instant_mode": false}`, want: "rerun"},
		{name: "env true", file: `{}`, env: "true", want: "auto"},
		{name: "env false", file: `{}`, env: "false", want: "rerun"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			isolateConfig(t)
			writeSettingsFile(t, tc.file)
			if tc.env != "" {
				t.Setenv("GOFUCK_INSTANT_MODE", tc.env)
			}
			settings, err := Load(nil)
			if err != nil {
				t.Fatal(err)
			}
			if settings.CaptureMode != tc.want {
				t.Fatalf("CaptureMode = %q, want %q", settings.CaptureMode, tc.want)
			}
		})
	}
}

func TestInvalidCaptureModeUsesAuto(t *testing.T) {
	isolateConfig(t)
	writeSettingsFile(t, `{"capture_mode": "sometimes"}`)
	settings, err := Load(nil)
	if err != nil {
		t.Fatal(err)
	}
	if settings.CaptureMode != "auto" {
		t.Fatalf("CaptureMode = %q, want auto", settings.CaptureMode)
	}
	if !strings.Contains(strings.Join(settings.InvalidValues, ","), "capture_mode sometimes") {
		t.Fatalf("InvalidValues = %v", settings.InvalidValues)
	}
}
