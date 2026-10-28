package logs

import (
	"strings"
	"testing"

	"github.com/leafevans/go-fuck/internal/conf"
)

func TestRenderBanner(t *testing.T) {
	buffer := resetLogsState(t)
	_ = buffer
	conf.Reset()

	info := VersionInfo{
		Version:     "1.2.3",
		GoVersion:   "go1.22",
		Shell:       "Bash 5.2",
		Platform:    "linux/amd64",
		Rules:       182,
		ConfigPath:  "/tmp/go-fuck/settings.json",
		InstantMode: "available (bash/zsh on Linux and macOS)",
		Fzf:         true,
		Alias:       "fuck",
	}

	banner := renderBanner(info)
	for _, want := range []string{
		"go-fuck",
		"1.2.3",
		"go1.22",
		"linux/amd64",
		"Bash 5.2",
		"182",
		"/tmp/go-fuck/settings.json",
		"available",
		"installed",
		"fuck",
		"https://github.com/leafevans/go-fuck",
	} {
		if !strings.Contains(banner, want) {
			t.Fatalf("banner does not contain %q:%s%s", want, newline, banner)
		}
	}
	if !strings.HasSuffix(banner, newline) {
		t.Fatal("banner does not end with a newline")
	}
}

func TestPrintVersionFallsBackToPlainLine(t *testing.T) {
	buffer := resetLogsState(t)
	PrintVersion(VersionInfo{Version: "1.2.3", GoVersion: "go1.22", Shell: "Bash 5.2"})
	if got, want := buffer.String(), "go-fuck 1.2.3 using Go go1.22 and Bash 5.2"+newline; got != want {
		t.Fatalf("PrintVersion() = %q, want %q", got, want)
	}
}
