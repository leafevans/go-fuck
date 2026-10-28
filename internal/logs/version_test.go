package logs

import (
	"strings"
	"testing"

	"github.com/leafevans/go-fuck/internal/conf"
)

// bannerInfo returns the fixture used by the banner tests.
func bannerInfo() VersionInfo {
	return VersionInfo{
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
}

func TestRenderBanner(t *testing.T) {
	_ = resetLogsState(t)
	conf.Reset()

	banner := renderBanner(bannerInfo(), false)
	for _, want := range []string{"go-fuck", "1.2.3", "go1.22", "linux/amd64", "Bash 5.2", "182 rules", "https://github.com/leafevans/go-fuck"} {
		if !strings.Contains(banner, want) {
			t.Fatalf("banner does not contain %q:%s%s", want, newline, banner)
		}
	}
	for _, unwanted := range []string{"/tmp/go-fuck/settings.json", "Instant", "fzf", "Alias"} {
		if strings.Contains(banner, unwanted) {
			t.Fatalf("compact banner should not contain %q:%s%s", unwanted, newline, banner)
		}
	}
	if !strings.HasSuffix(banner, newline) {
		t.Fatal("banner does not end with a newline")
	}
}

func TestRenderVerboseBanner(t *testing.T) {
	_ = resetLogsState(t)
	conf.Reset()

	banner := renderBanner(bannerInfo(), true)
	for _, want := range []string{"/tmp/go-fuck/settings.json", "Instant", "available", "fzf", "installed", "Alias", "fuck"} {
		if !strings.Contains(banner, want) {
			t.Fatalf("verbose banner does not contain %q:%s%s", want, newline, banner)
		}
	}
}

func TestPrintVersionFallsBackToPlainLine(t *testing.T) {
	buffer := resetLogsState(t)
	PrintVersion(VersionInfo{Version: "1.2.3", GoVersion: "go1.22", Shell: "Bash 5.2"}, false)
	if got, want := buffer.String(), "go-fuck 1.2.3 using Go go1.22 and Bash 5.2"+newline; got != want {
		t.Fatalf("PrintVersion() = %q, want %q", got, want)
	}
}
