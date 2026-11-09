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

// TestOutputIsTerminal checks that a test buffer is not reported as a
// terminal, which lets callers skip the interactive version banner work.
func TestOutputIsTerminal(t *testing.T) {
	_ = resetLogsState(t)
	if OutputIsTerminal() {
		t.Fatal("OutputIsTerminal() = true for a buffer, want false")
	}
}

// setBannerEnvironment isolates the environment variables that influence the
// color profile detection.
func setBannerEnvironment(t *testing.T, term, colorTerm, wtSession, noColor string) {
	t.Helper()
	t.Setenv("TERM", term)
	t.Setenv("COLORTERM", colorTerm)
	t.Setenv("WT_SESSION", wtSession)
	t.Setenv("NO_COLOR", noColor)
}

func TestRenderBannerTrueColor(t *testing.T) {
	_ = resetLogsState(t)
	conf.Reset()
	setBannerEnvironment(t, "xterm-256color", "truecolor", "", "")

	banner := renderBanner(bannerInfo(), false)
	if !strings.Contains(banner, "\x1b[38;2;") {
		t.Fatalf("banner does not contain 24 bit color sequences")
	}
}

func TestRenderBanner256Colors(t *testing.T) {
	_ = resetLogsState(t)
	conf.Reset()
	setBannerEnvironment(t, "xterm-256color", "", "", "")

	banner := renderBanner(bannerInfo(), false)
	if !strings.Contains(banner, "\x1b[38;5;") {
		t.Fatalf("banner does not contain 256 color sequences")
	}
	if strings.Contains(banner, "\x1b[38;2;") {
		t.Fatalf("256 color banner must not use 24 bit colors")
	}
}

func TestRenderBannerBasicColors(t *testing.T) {
	_ = resetLogsState(t)
	conf.Reset()
	setBannerEnvironment(t, "xterm", "", "", "")

	banner := renderBanner(bannerInfo(), false)
	if !strings.Contains(banner, foreRed) {
		t.Fatalf("banner does not use the 16 color fallback")
	}
	if strings.Contains(banner, "\x1b[38;") {
		t.Fatalf("16 color banner must not use extended color sequences")
	}
}

func TestRenderBannerNoColorEnvironment(t *testing.T) {
	_ = resetLogsState(t)
	conf.Reset()
	setBannerEnvironment(t, "xterm-256color", "truecolor", "", "1")

	banner := renderBanner(bannerInfo(), false)
	if strings.Contains(banner, "\x1b[") {
		t.Fatalf("NO_COLOR banner contains escape sequences")
	}
	if !strings.Contains(banner, "182 rules") {
		t.Fatalf("NO_COLOR banner lost its text")
	}
}

func TestRenderBannerNoColorSetting(t *testing.T) {
	_ = resetLogsState(t)
	conf.Reset()
	setBannerEnvironment(t, "xterm-256color", "truecolor", "", "")
	conf.Current().NoColors = true

	banner := renderBanner(bannerInfo(), false)
	if strings.Contains(banner, "\x1b[") {
		t.Fatalf("no_colors banner contains escape sequences")
	}
}
